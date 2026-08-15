# Commerce's own image.
#
# Two binaries: this product's server — the platform plus its four modules — and
# the platform's migration runner, built from the exact core version go.mod
# pins. The migrations come from the same place, out of the module cache, so the
# schema an image applies is the schema the code in it was written against.
# Taking them from anywhere else is how a deployment ends up one release ahead
# of its own database.

FROM golang:1.26-alpine AS build

RUN apk add --no-cache git ca-certificates

WORKDIR /src

# Dependencies first, so a change to our own code does not re-download the core.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# What this build tells an app store it is.
#
# Derived from the core version go.mod pins rather than passed in, because it is
# not a choice: PlatformVersion is what every manifest's `"platform": ">=1.0.0"`
# is checked against, and the honest answer is the version of the platform
# actually compiled into this image. Left to a build argument it defaults to
# 0.0.0, and the instance then refuses every manifest it ships with — it cannot
# read its own catalogue, so it will not start.
RUN core=$(go list -m -f '{{.Version}}' github.com/gerege-systems/open-gerege-nexus/backend) \
    && version=${core#v} \
    && echo "building as platform ${version}" \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags="-w -s -X github.com/gerege-systems/open-gerege-nexus/backend/internal/platform.PlatformVersion=${version}" \
    -o /bin/commerce .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-w -s" \
    -o /bin/migrate github.com/gerege-systems/open-gerege-nexus/backend/cmd/migrate

# The platform's migration files, which are data in the module rather than code.
# The path carries the version, so this reads it back rather than guessing.
RUN core=$(go list -m -f '{{.Dir}}' github.com/gerege-systems/open-gerege-nexus/backend) \
    && mkdir -p /out/db && cp -r "$core/db/migrations" /out/db/migrations \
    && chmod -R u+w /out/db

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata wget \
    && adduser -D -u 10001 commerce

WORKDIR /app

COPY --from=build /bin/commerce /app/commerce
COPY --from=build /bin/migrate  /app/migrate
COPY --from=build /out/db       /app/db
# This product's own migrations, which are not the platform's: five tables the
# commerce split had to re-declare because the originals were tangled into
# migrations every deployment had already run. Applied with their own
# MIGRATIONS_TABLE — see the compose file.
COPY migrations                 /app/db/commerce
# The catalogue this instance offers: the platform's apps as the pinned core
# ships them, and this product's four beside them.
COPY catalog                    /app/catalog

USER commerce
EXPOSE 8080

CMD ["/app/commerce"]
