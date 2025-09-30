FROM golang:1-bookworm AS build

RUN apt-get update && \
	apt-get install -y npm libpcap-dev

ADD . /src
WORKDIR /src

RUN cd web/ui && \
	rm -Rf node_modules && \
	npm i && \
	npm run build && \
	cd ../..

RUN cd web/admin && \
	rm -Rf node_modules && \
	npm i && \
	npm run build && \
	cd ../..

RUN go install github.com/swaggo/swag/cmd/swag@latest && \
	swag i --exclude ./web/ui --output web/docs && \
	go build -trimpath -ldflags="-s -w \
	-X=github.com/sensepost/gowitness/internal/version.GitHash=$(git rev-parse --short HEAD) \
	-X=github.com/sensepost/gowitness/internal/version.GoBuildEnv=$(go version | cut -d' ' -f 3,4 | sed 's/ /_/g') \
	-X=github.com/sensepost/gowitness/internal/version.GoBuildTime=$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
	-o gowitness

# Install naabu port scanner
RUN go install -v github.com/projectdiscovery/naabu/v2/cmd/naabu@latest

FROM ghcr.io/go-rod/rod

# Install runtime dependencies for naabu
RUN apt-get update && apt-get install -y libpcap0.8 && rm -rf /var/lib/apt/lists/*

COPY --from=build /src/gowitness /usr/local/bin/gowitness
COPY --from=build /src/web/admin/dist /app/web/admin/dist
COPY --from=build /go/bin/naabu /usr/local/bin/naabu

# Create app directory and set as working directory
RUN mkdir -p /app
WORKDIR /app

# Create symlink for scan run command compatibility
RUN ln -sf /usr/local/bin/gowitness /app/gowitness

# Create directories for project data
RUN mkdir -p /app/targets /app/projects /app/nginx-config

EXPOSE 7171 8080

VOLUME ["/data", "/app/targets", "/app/projects", "/app/nginx-config"]

ENTRYPOINT ["dumb-init", "--"]