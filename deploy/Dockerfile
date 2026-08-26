# Gerege Business's own image.
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
#
# Which package holds it is asked rather than assumed. PlatformVersion moved
# from internal/platform to internal/kernel/config after the core's v1.13.0,
# and `-X` naming a symbol that does not exist is **silently ignored** by the
# linker: a stale path builds cleanly, produces an image that reports the
# default 1.1.0, and that image then refuses every manifest it ships with and
# never comes up. Nothing in the build log says so. So the two candidates are
# probed in order and a core with neither fails the build here, loudly.
RUN core=$(go list -m -f '{{.Version}}' github.com/gerege-systems/open-gerege-nexus/backend) \
    && version=${core#v} \
    && base=github.com/gerege-systems/open-gerege-nexus/backend/internal \
    && for pkg in kernel/config platform; do \
         if go list "$base/$pkg" >/dev/null 2>&1; then sym="$base/$pkg.PlatformVersion"; break; fi; \
       done \
    && { [ -n "${sym:-}" ] || { echo "no PlatformVersion package in ${core}"; exit 1; }; } \
    && echo "building as platform ${version} (${sym##*/internal/})" \
    && CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags="-w -s -X ${sym}=${version}" \
    -o /bin/business .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-w -s" \
    -o /bin/migrate github.com/gerege-systems/open-gerege-nexus/backend/cmd/migrate

# The platform's migration files, which are data in the module rather than code.
# The path carries the version, so this reads it back rather than guessing.
RUN core=$(go list -m -f '{{.Dir}}' github.com/gerege-systems/open-gerege-nexus/backend) \
    && mkdir -p /out/db && cp -r "$core/db/migrations" /out/db/migrations \
    && chmod -R u+w /out/db

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata wget \
    && adduser -D -u 10001 business

WORKDIR /app

COPY --from=build /bin/business /app/business
COPY --from=build /bin/migrate  /app/migrate
COPY --from=build /out/db       /app/db
# This product's own migrations, which are not the platform's: five tables the
# split had to re-declare because the originals were tangled into
# migrations every deployment had already run. Applied with their own
# MIGRATIONS_TABLE — see the compose file.
COPY migrations                 /app/db/business
# The catalogue this instance offers: the platform's apps as the pinned core
# ships them, and this product's four beside them.
COPY catalog                    /app/catalog

USER business
EXPOSE 8080

CMD ["/app/business"]
