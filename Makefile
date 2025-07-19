copy:
	@aws s3 cp ./dist/index.html s3://$(bucket)/ \
		--content-type "text/html" \
		--cache-control "no-cache, no-store, must-revalidate"

sync:
	@aws s3 sync ./dist s3://$(bucket)/ --exclude index.html --exclude "images/*" --delete

build:
	@npm run build

deploy: build copy sync

server:
	@npm run dev