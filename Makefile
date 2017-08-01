rversion := 0.6.7

.PHONY: build crelease cbuild cpush

release : bbuild cbuild cpush

bbuild:
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.releaseVersion=$(rversion)" -o ./resorcerer

cbuild :
	docker build --build-arg rversion=$(rversion) -t quay.io/mhausenblas/resorcerer:$(rversion) .

cpush :
	docker push quay.io/mhausenblas/resorcerer:$(rversion)
	@rm ./resorcerer
