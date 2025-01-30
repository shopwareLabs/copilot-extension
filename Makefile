fetch-data:
	rm -rf data
	git clone --depth=1 https://github.com/shopware/shopware.git data
	composer install -d data --no-scripts
	./data/bin/console list --format json > data/commands.json
	node cli-console-markdown.mjs > all-commands.md
	cd data && rm -rf .git && find . -mindepth 1 ! -regex '^./src.*' -delete
	rm -rf data/src/WebInstaller
	rm -f data/src/Administration/{README.md,LICENSE}
	rm -rf data/src/Administration/Resources/app/administration/build
	rm -rf data/src/Administration/Resources/app/administration/eslint-rules
	rm -rf data/src/Administration/Resources/app/administration/patches
	rm -rf data/src/Administration/Resources/app/administration/scripts
	rm -rf data/src/Administration/Resources/app/administration/static
	rm -rf data/src/Administration/Resources/app/administration/test
	rm -rf data/src/Administration/Resources/app/administration/*.js
	find data/src/Administration -name "*.spec.js" -o -name "*.spec.ts" -delete
	rm data/src/Core/locales.php
	rm -rf data/src/Storefront/Resources/app/storefront/test
	rm -rf data/src/Storefront/Resources/app/storefront/static/draco
	git clone --depth=1 https://github.com/shopware/docs.git data/docs
	cd data/docs && rm -rf .git && cd ..
	find data/docs -maxdepth 1 -type f -delete
	rm -rf data/docs/.github data/docs/.vscode data/docs/assets
	git clone --depth=1 https://github.com/shopware/frontends.git data/frontends_tmp
	mv data/frontends_tmp/apps/docs/src data/frontends
	rm -rf data/frontends_tmp
	rm -rf data/frontends/ai data/frontends/.assets data/frontends/public
	mv all-commands.md data/docs/

