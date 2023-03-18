FROM golang:1.20.1-bullseye

WORKDIR /postParser

COPY . .

EXPOSE 443 3000 3100

RUN go get -d -v ./... \
		&& go get -v -u golang.org/x/tools//gopls \
		&& git clone https://github.com/go-delve/delve \
            && cd delve \
            && go install github.com/go-delve/delve/cmd/dlv \
		&& go get -v -u github.com/ramya-rao-a/go-outline \
		&& go get -v -u github.com/cweill/gotests \
		&& go get -v -u github.com/haya14busa/goplay \
		&& go get -v -u github.com/fatih/gomodifytags \
		&& go get -v -u github.com/josharian/impl \
		&& go get -v -u github.com/cweill/gotests