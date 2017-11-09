docker container run --rm -w /go/src/lifestyle -v $(pwd):/go/src/lifestyle golang:1.9.0-alpine3.6 go test -coverprofile=a.out lifestyle/fav
docker container run --rm -w /go/src/lifestyle -v $(pwd):/go/src/lifestyle golang:1.9.0-alpine3.6 go test -coverprofile=b.out lifestyle/somewhere && sed '1d' b.out >> a.out && sed -i.bak 's/lifestyle/./g' a.out && go tool cover -html=a.out
