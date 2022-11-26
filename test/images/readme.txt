These images are exported as *.tars to make testing easier and less reliant on whatever is stored in a registry.
In most cases, testing with the *.tar and the image from a registry must provide the same result - only parsing could differ.

In order to save an image like this, use
    docker save <image> -o <path_to_tar>
