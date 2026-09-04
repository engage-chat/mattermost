ifneq ($(origin CUSTOMIZE_SOURCE_DIR), undefined)
  $(error CUSTOMIZE_SOURCE_DIR is already set (origin=$(origin CUSTOMIZE_SOURCE_DIR)))
endif

CUSTOMIZE_SOURCE_DIR = $(BUILD_WEBAPP_DIR)/channels/dist

customize-assets:
	@echo "🚀 Starting customize-assets"
	@echo "CUSTOM_SERVICE_NAME = $(CUSTOM_SERVICE_NAME)"
	@echo "CUSTOM_PLATFORM_NAME = $(CUSTOM_PLATFORM_NAME)"
	@echo "CUSTOM_JP_PLATFORM_NAME = $(CUSTOM_JP_PLATFORM_NAME)"
	@echo "CUSTOMIZE_SOURCE_DIR = $(CUSTOMIZE_SOURCE_DIR)"

	@echo "replacing service and platform names in i18n files..."
	sed -i'' -e '/"about\.notice"/!{ /"about\.copyright"/!s/Mattermost/$(CUSTOM_JP_PLATFORM_NAME)/g; }' $(CUSTOMIZE_SOURCE_DIR)/i18n/ja.*.json
	sed -i'' -e 's/GitLab/$(CUSTOM_SERVICE_NAME)/g' -e 's/{service}/$(CUSTOM_SERVICE_NAME)/g' -e '/"about\.notice"/!{ /"about\.copyright"/!s/Mattermost/$(CUSTOM_PLATFORM_NAME)/g; }' $(CUSTOMIZE_SOURCE_DIR)/i18n/*.json
	sed -i'' -e 's/Mattermost/$(CUSTOM_JP_PLATFORM_NAME)/g' i18n/ja.json
	sed -i'' -e 's/{{.Service}}/$(CUSTOM_SERVICE_NAME)/g' -e 's/Mattermost/$(CUSTOM_PLATFORM_NAME)/g' i18n/*.json

	@echo "removing GitLab icon from login screen..."
	icon_str='"svg",\{width:"[0-9]+",height:"[0-9]+",viewBox:"0 0 [0-9]+ [0-9]+",fill:"none",xmlns:"http:\/\/www\.w3\.org\/2000\/svg","aria-label":t\(\{id:"generic_icons\.login\.gitlab",defaultMessage:"Gitlab Icon"\}\)\}'; \
	echo "icon_str: $${icon_str}"; \
	grep -l "generic_icons.login.gitlab" $(CUSTOMIZE_SOURCE_DIR)/*.js | while read -r file; do \
		if [ -n "$${file}" ]; then \
			echo "-> Found file: $${file}. Modifying content..."; \
			sed -i'' -E \
				-e "s|$${icon_str}|\"span\",\{\}|g" \
				-e 's/external-login-button-label//g' \
				"$${file}"; \
			if grep -q -E "$${icon_str}" "$${file}"; then \
				echo "::error title=Removing GitLab icon Verification Error::Failed to replace GitLab icon in $${file}."; \
				exit 1; \
			fi; \
		else \
			echo "::error title=Removing GitLab icon Error::GitLab icon pattern not found."; \
			exit 1; \
		fi; \
	done;

	@echo "hiding Mattermost logo at the top left..."
	@files=$$(grep -l "hfroute-header" $(CUSTOMIZE_SOURCE_DIR)/*.js 2>/dev/null); \
	if [ -z "$${files}" ]; then \
		echo "::error title=Hiding Mattermost logo Error::hfroute-header pattern not found in any JS file. Upstream code might have changed."; \
		exit 1; \
	fi; \
	for file in $${files}; do \
		echo "-> Found JS file: $${file}. Modifying content..."; \
		sed -i'' -E 's/(className:[^}]*hfroute-header)/style:{visibility:"hidden"},\1/g' "$${file}"; \
		if ! grep -q 'style:{visibility:"hidden"}[^}]*hfroute-header' "$${file}"; then \
			echo "::error title=Hiding Mattermost logo Verification Error::Failed to replace hfroute-header in $${file}. Upstream code might have changed."; \
			exit 1; \
		fi; \
	done

	@echo "hiding ID/password login form..."
	grep -l "login-body-card-form" $(CUSTOMIZE_SOURCE_DIR)/*.css | while read -r css_file; do \
		if [ -n "$${css_file}" ]; then \
			echo "-> Found file: $${css_file}. Appending login form hiding rules..."; \
			echo ".login-body-card-form { display: none !important; } .login-body-card-form-divider { display: none !important; } .login-body-alternate-link { display: none !important; }" >> "$${css_file}"; \
			if ! grep -q "\.login-body-card-form { display: none !important; }" "$${css_file}"; then \
				echo "::error title=Hiding login form Verification Error::Failed to append rules to $${css_file}."; \
				exit 1; \
			fi; \
		else \
			echo "::error title=Hiding login form Error::login-body-card-form pattern not found in any CSS file."; \
			exit 1; \
		fi; \
	done

	@echo "hiding loading screen icon..."
	@loading_css_files=$$(grep -l "LoadingAnimation__compass" $(CUSTOMIZE_SOURCE_DIR)/*.css 2>/dev/null); \
	if [ -z "$${loading_css_files}" ]; then \
		echo "::error title=Hiding loading screen Error:: .LoadingAnimation__compass pattern not found in any CSS file. Upstream code might have changed."; \
		exit 1; \
	fi; \
	for target_css in $${loading_css_files}; do \
		echo "-> Appending loading screen hiding rules to $${target_css}..."; \
		echo ".LoadingAnimation__compass { display: none !important; }" >> "$${target_css}"; \
		if ! grep -q "\.LoadingAnimation__compass { display: none !important; }" "$${target_css}"; then \
			echo "::error title=Hiding loading screen Verification Error::Failed to append rules to $${target_css}."; \
			exit 1; \
		fi; \
	done

	@echo "✅ Completed customize-assets"
