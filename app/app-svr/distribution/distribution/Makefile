BAPI_BRANCH ?= master

proto: clean
	mkdir -p ./bapis/bilibili/app/
	git clone git@git.bilibili.co:bapis/bapis.git -b ${BAPI_BRANCH} bapis-tmp
	cp -r bapis-tmp/bilibili/app/distribution ./bapis/bilibili/app
	cp -f fixture/fixture.proto ./bapis/fixture.proto
	cp -f bapis-tmp/bilibili/app/distribution/service.proto ./api/service.proto
	cp -r bapis-tmp/bilibili/extension ./bapis/bilibili/extension
	rm -rf bapis-tmp
	find bapis -name "buf.yaml" | xargs rm -rf
	statik -f -Z -ns=deivcesetting -src=bapis -dest=. -include=*.proto -p=embed
	gofmt -w -s ./embed
	gsed -i 's/git.bilibili.co\/bapis\/bapis-go\/bilibili\/app\/distribution\;distribution/api/g' ./api/service.proto
	gsed -i 's/import "bilibili\/extension\/crane\/crane.proto";/\/\/import "bilibili\/extension\/crane\/crane.proto";/' ./api/service.proto
	gsed -i 's/option (bilibili.extension.crane.bili_grpc_host) = "main.grpc.bilibili.com";/\/\/option (bilibili.extension.crane.bili_grpc_host) = "main.grpc.bilibili.com";/' ./api/service.proto
	go generate -v ./...

clean:
	rm -rf bapis bapis-tmp distribution