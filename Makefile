rversion := 0.6.18

.PHONY: build crelease cbuild cpush

release : bbuild cbuild cpush

bbuild:
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.releaseVersion=$(rversion)" -o ./resorcerer

cbuild :
	docker build --build-arg rversion=$(rversion) -t quay.io/mhausenblas/resorcerer:$(rversion) .
	# cd deployments && docker build -t quay.io/mhausenblas/loadgenerator:0.2 .

cpush :
	docker push quay.io/mhausenblas/resorcerer:$(rversion)
	# docker push quay.io/mhausenblas/loadgenerator:0.2
	@rm ./resorcerer
