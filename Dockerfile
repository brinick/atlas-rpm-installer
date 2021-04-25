FROM atlas/centos7-atlasos:latest

RUN wget https://golang.org/dl/go1.15.5.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go1.15.5.linux-amd64.tar.gz \
    && export PATH=$PATH:/usr/local/go/bin \
    && go version

CMD export PATH=$PATH:/usr/local/go/bin
