FROM scratch

# L1 - copy the executable in 1 path
ADD app /executable

# L2 - copy the executable in another path
ADD app /executable_2

# L3 - ask to delete the executable from L2
RUN ["/executable", "delete", "/executable_2"]

# L4 - copy the executable in yet another path
ADD app /executable_3

# L5 - ask to change the L4 executable (note: this invalidates the actual file)
RUN ["/executable", "change", "/executable_3"]

# L6 - empty layer that modifes an env variable
ENV test=/test