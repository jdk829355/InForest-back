archs=(amd64 arm64)

for arch in ${archs[@]}
do
	env GOOS=linux GOARCH=${arch} go build -o ./bin/server_${arch} ./cmd/server
done