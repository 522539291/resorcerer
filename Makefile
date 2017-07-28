rversion := 0.1.0

.PHONY: build crelease cbuild cpush

crelease : bbuild cbuild cpush

bbuild:
	GOOS=linux GOARCH=amd64 go build -o ./resorcerer

cbuild :
	docker build --build-arg rversion=$(rversion) -t quay.io/mhausenblas/resorcerer:$(rversion) .

cpush :
	docker push quay.io/mhausenblas/resorcerer:$(rversion)
	@rm ./resorcerer
