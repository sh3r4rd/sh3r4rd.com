copy.aws:
	@aws s3 cp ./dist/index.html s3://sh3r4rd.com/ \
		--content-type "text/html" \
		--cache-control "no-cache, no-store, must-revalidate"

sync.aws:
	@aws s3 sync ./dist s3://sh3r4rd.com --exclude index.html --exclude "images/*" --delete