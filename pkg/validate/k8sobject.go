package validate

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"gitlab.bit9.local/octarine/cbctl/internal/util/httptool"
	"gitlab.bit9.local/octarine/cbctl/pkg/model/resource"
	"sigs.k8s.io/yaml"
)

const (
	validateResourceTemplate = "%s/guardrails/validator/%s/resource"

	yamlSeparator     = "\n---"
	numberOfConsumers = 50
)

var yamlExtensions = []string{".yaml", ".yml"}

// K8SObjectHandler has all the fields for sending request to validator service.
type K8SObjectHandler struct {
	session   *httptool.RequestSession
	basePath  string
	buildStep string
	namespace string
	path      string
}

// NewK8SObjectValidateHandler will create a handler for validate cmd.
func NewK8SObjectValidateHandler(saasTmpl, orgKey, apiID, apiKey, buildStep, namespace, path string) *K8SObjectHandler {
	basePath := fmt.Sprintf("%s/v1/orgs/%s", strings.Trim(saasTmpl, "/"), orgKey)
	session := httptool.NewRequestSession(apiID, apiKey)

	return newK8SObjectValidateHandler(session, basePath, buildStep, namespace, path)
}

func newK8SObjectValidateHandler(session *httptool.RequestSession,
	basePath, buildStep, namespace, path string) *K8SObjectHandler {
	return &K8SObjectHandler{
		session:   session,
		basePath:  basePath,
		buildStep: buildStep,
		namespace: namespace,
		path:      path,
	}
}

func (h K8SObjectHandler) hasYamlExtension(path string) bool {
	for _, ext := range yamlExtensions {
		if filepath.Ext(path) == ext {
			return true
		}
	}

	return false
}

func (h K8SObjectHandler) readResourceFromStdin() ([]byte, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return nil, fmt.Errorf("can't get information on stdin %v", err)
	}

	if info.Mode()&os.ModeCharDevice != 0 || info.Size() <= 0 {
		return nil, fmt.Errorf("the command is intended to work with pipes")
	}

	reader := bufio.NewReader(os.Stdin)

	var output []rune

	for {
		input, _, err := reader.ReadRune()
		if err != nil && err == io.EOF {
			break
		}

		output = append(output, input)
	}

	return []byte(string(output)), nil
}

func (h K8SObjectHandler) separateYaml(data []byte) ([]string, error) {
	yamls := strings.Split(string(data), yamlSeparator)
	result := make([]string, 0)

	for _, y := range yamls {
		var temp interface{}

		if strings.TrimSpace(y) == "" {
			continue
		}

		err := yaml.Unmarshal([]byte(y), &temp)
		if err != nil {
			return nil, err
		}

		result = append(result, y)
	}

	return result, nil
}

func (h K8SObjectHandler) produceValidationJobsForMultiYaml(
	multiYaml []byte, filePath string, producedJobs chan<- *Job,
) {
	var resourcesData []string
	if result, err := h.separateYaml(multiYaml); err == nil {
		resourcesData = result
	} else {
		resourcesData = []string{string(multiYaml)}
	}

	for _, data := range resourcesData {
		producedJobs <- NewJob(data, filePath)
	}
}

func (h K8SObjectHandler) handleDir(dirPath string, producedJobs chan<- *Job) {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		producedJobs <- NewJobWithError(fmt.Sprintf("PATH ERROR %v: can't get folder's files", dirPath), dirPath)
		return
	}

	for _, f := range files {
		h.handlePath(path.Join(dirPath, f.Name()), producedJobs)
	}
}

func (h K8SObjectHandler) handlePath(path string, producedJobs chan<- *Job) {
	pathInfo, err := os.Stat(path)
	if err != nil {
		producedJobs <- NewJobWithError("can't get information about the path", path)
		return
	}

	switch {
	case pathInfo.Mode().IsDir():
		h.handleDir(path, producedJobs)
	case pathInfo.Mode().IsRegular() && h.hasYamlExtension(pathInfo.Name()):
		if data, err := ioutil.ReadFile(path); err != nil {
			producedJobs <- NewJobWithError(fmt.Sprintf("can't read from file (%v)", err), path)
		} else {
			var obj interface{}
			if err := yaml.Unmarshal(data, &obj); err != nil {
				producedJobs <- NewJobWithError(fmt.Sprintf("invalid yaml file (%v)", err), path)
			} else {
				h.produceValidationJobsForMultiYaml(data, path, producedJobs)
			}
		}
	}
}

func (h K8SObjectHandler) checkFilesAmount(path string) error {
	fileCount := 0
	countFileFunc := func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && h.hasYamlExtension(filePath) {
			fileCount++
		}

		return nil
	}

	if err := filepath.Walk(path, countFileFunc); err != nil {
		return err
	}

	if fileCount == 0 {
		return fmt.Errorf("found 0 yaml files to validate (%s)", path)
	}

	return nil
}

func (h K8SObjectHandler) produceValidationJobs(producedJobs chan<- *Job) error {
	if h.path == "-" {
		data, err := h.readResourceFromStdin()
		if err != nil {
			return err
		}

		h.produceValidationJobsForMultiYaml(data, "STDIN", producedJobs)
	} else {
		if err := h.checkFilesAmount(h.path); err != nil {
			return err
		}

		h.handlePath(h.path, producedJobs)
	}

	return nil
}

func (h K8SObjectHandler) consumeValidationJobs(
	consumersGroup *sync.WaitGroup, producedJobs <-chan *Job, consumedJobs chan<- *Job) {
	defer consumersGroup.Done()

	for job := range producedJobs {
		if job.error != "" {
			consumedJobs <- job
			continue
		}

		resourceResult, err := h.getResourceViolations(job.resourceData)
		if err != nil {
			job.error = fmt.Sprintf("failed get violation: %s", err.Error())
			consumedJobs <- job

			continue
		}

		job.result = resourceResult
		consumedJobs <- job
	}
}

func (h K8SObjectHandler) analyzeResults(
	analyzerGroup *sync.WaitGroup, consumedJobs <-chan *Job, result *resource.ValidatedResources) {
	defer analyzerGroup.Done()

	for job := range consumedJobs {
		if job.error != "" {
			result.Errors = append(result.Errors, fmt.Sprintf("%v (%v)", job.error, job.filePath))
		} else if job.result != nil {
			result.ViolatedResources = append(result.ViolatedResources, resource.ValidatedResource{
				Scope:            job.result.Scope,
				FilePath:         job.filePath,
				Policy:           job.result.Policy,
				PolicyViolations: job.result.PolicyViolations,
			})
		}
	}
}

// Validate will will validate the resource using the validate API.
func (h K8SObjectHandler) Validate() (resource.ValidatedResources, error) {
	var consumersGroup, analyzerGroup sync.WaitGroup

	producedJobs := make(chan *Job)
	consumedJobs := make(chan *Job)

	// Start producing:
	var producingError error

	go func() {
		producingError = h.produceValidationJobs(producedJobs)
		close(producedJobs)
	}()

	// Start consuming:
	for i := 0; i < numberOfConsumers; i++ {
		consumersGroup.Add(1)

		go h.consumeValidationJobs(&consumersGroup, producedJobs, consumedJobs)
	}

	// Start analyzing:
	analyzerGroup.Add(1)

	var result resource.ValidatedResources

	go h.analyzeResults(&analyzerGroup, consumedJobs, &result)

	consumersGroup.Wait()

	close(consumedJobs)

	analyzerGroup.Wait()

	if producingError != nil {
		return resource.ValidatedResources{}, producingError
	}

	return result, nil
}

func (h K8SObjectHandler) getResourceViolations(resourceData string) (*resource.ValidatedResourceResponse, error) {
	if err := CheckValidBuildStep(h.buildStep); err != nil {
		return nil, err
	}

	validateURL, err := url.Parse(fmt.Sprintf(validateResourceTemplate, h.basePath, h.buildStep))
	if err != nil {
		return nil, fmt.Errorf("unexpected error, failed to parse validate url: %v", err)
	}

	if h.namespace != "" {
		params := url.Values{}
		params.Add("namespace", h.namespace)
		validateURL.RawQuery = params.Encode()
	}

	type ResourceViolationsPayload struct {
		ResourceData []byte `json:"resource_data"`
	}

	payload := ResourceViolationsPayload{
		ResourceData: []byte(resourceData),
	}

	statusCode, resp, err := h.session.RequestData(http.MethodPost, validateURL.String(), payload)
	if err != nil {
		if errMessage, ok := httptool.TryReadErrorResponse(resp); ok {
			return nil, fmt.Errorf("failed get violations from backend: %s", errMessage)
		}

		return nil, fmt.Errorf("failed get violations from backend, status code: %d", statusCode)
	}

	var result *resource.ValidatedResourceResponse

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return result, nil
}
