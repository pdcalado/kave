#!/usr/bin/env bash

usage() {
    echo -e "This script generates tools that you can include in your "'$PATH'" to easily invoke kave using the docker image (kave-get, kave-set, kave-token).\n"
    echo -e "The generated tools assume you've got an env file somewhere with the AUTH0 env vars: AUTH0_AUDIENCE,AUTH0_CLIENT_ID,AUTH0_DOMAIN,AUTH0_CLIENT_SECRET.\n"
    echo -e "The generated tools also assume that tokens are cached in ~/.cache/kave/token.\n"
    echo -e "Usage:\n$0 <path-to-env-file>\n"
}

if [ $# -lt 1 ]
then
    usage
    exit 1
fi

cat << EOF > ./bin/kave-token && chmod +x ./bin/kave-token
#!/usr/bin/env bash

set -euf -o pipefail

source $1

docker run --rm --env-file $1 ghcr.io/pdcalado/kave:latest \\
    kave token \$AUTH0_CLIENT_ID
EOF

cat << EOF > ./bin/kave-refresh-token && chmod +x ./bin/kave-refresh-token
#!/usr/bin/env bash

set -euf -o pipefail

source $1

mkdir -p ~/.cache/kave

docker run --rm --env-file $1 ghcr.io/pdcalado/kave:latest \\
    kave token \$AUTH0_CLIENT_ID > ~/.cache/kave/token
EOF

cat << EOF > ./bin/kave-get && chmod +x ./bin/kave-get
#!/usr/bin/env bash

set -euf -o pipefail

source $1

docker run --rm --env-file $1 ghcr.io/pdcalado/kave:latest \\
    kave --token \$(cat ~/.cache/kave/token) --url \$AUTH0_AUDIENCE get "\$1"
EOF

cat << EOF > ./bin/kave-set && chmod +x ./bin/kave-set
#!/usr/bin/env bash

set -euf -o pipefail

source $1

docker run --rm --env-file $1 ghcr.io/pdcalado/kave:latest \\
    kave --token \$(cat ~/.cache/kave/token) --url \$AUTH0_AUDIENCE set "\$1" "\$2"
EOF