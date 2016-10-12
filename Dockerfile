FROM bongo227/llvmgo

# Set gopath
ENV GOPATH /go

# Fix locale
RUN locale-gen en_US.UTF-8
ENV LC_ALL "en_US.UTF-8"
ENV LANG "en_US.UTF-8"
ENV LANGUAGE "en_US.UTF-8"

# Install llvm
RUN go install llvm.org/llvm/bindings/go/llvm

# Add usr/local/lib to linker path
RUN echo "/go/src/llvm.org/llvm/bindings/go/llvm/workdir/llvm_build/lib\n" >> /etc/ld.so.conf
RUN ldconfig


# Install bash testing library
RUN git clone https://github.com/sstephenson/bats.git
RUN ./bats/install.sh /usr/local

# Install lli
RUN apt-get install llvm-runtime -y

WORKDIR /go/src/bitbucket.com/bongo227/furlang

# Build the compiler
# COPY . .
# RUN ["go", "build", "-tags='llvm'", "-o=furlang", "compiler.go"]

# Setup compiler command
ENTRYPOINT ["bats"]
CMD ["examples.bats"]