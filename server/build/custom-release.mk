# ==============================================================================
# Caching Constraints for Asset Customization
# ==============================================================================
# Static assets in dist/ (JS, CSS, ..etc) are served with long-term caching
# headers (Cache-Control: max-age=31556926, public).
#
# Because content hashes in filenames are determined during the webpack build,
# post-build modifications in this Makefile DO NOT change the asset filenames.
#
# NOTE:
# Take note of browser caching behavior when editing this file.
# For example, during version upgrades or updates, if component code (JS) changes
# but CSS source does not, modifying CSS here will NOT bust the CSS cache, leading
# to stale styles on clients.
# ==============================================================================

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
	@icon_str='"svg",\{width:"[0-9]+",height:"[0-9]+",viewBox:"0 0 [0-9]+ [0-9]+",fill:"none",xmlns:"http:\/\/www\.w3\.org\/2000\/svg","aria-label":t\(\{id:"generic_icons\.login\.gitlab",defaultMessage:"Gitlab Icon"\}\)\}'; \
	echo "icon_str: $${icon_str}"; \
	gitlab_files=$$(grep -l 'id:"generic_icons\.login\.gitlab"' $(CUSTOMIZE_SOURCE_DIR)/*.js 2>/dev/null); \
	if [ -z "$${gitlab_files}" ]; then \
		echo "::error title=Removing GitLab icon Error::GitLab icon pattern not found in any JS file. Upstream code might have changed."; \
		exit 1; \
	fi; \
	for file in $${gitlab_files}; do \
		echo "-> Found file: $${file}. Modifying content..."; \
		sed -i'' -E \
			-e "s|$${icon_str}|\"span\",\{\}|g" \
			-e 's/external-login-button-label//g' \
			"$${file}"; \
		if grep -q 'id:"generic_icons\.login\.gitlab"' "$${file}"; then \
			echo "::error title=Removing GitLab icon Verification Error::Failed to replace GitLab icon in $${file}. Upstream code might have changed."; \
			exit 1; \
		fi; \
	done

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
	@login_css_files=$$(grep -l "login-body-card-form" $(CUSTOMIZE_SOURCE_DIR)/*.css 2>/dev/null); \
	if [ -z "$${login_css_files}" ]; then \
		echo "::error title=Hiding login form Error::login-body-card-form pattern not found in any CSS file. Upstream code might have changed."; \
		exit 1; \
	fi; \
	for css_file in $${login_css_files}; do \
		echo "-> Found file: $${css_file}. Appending login form hiding rules..."; \
		echo ".login-body-card-form { display: none !important; } .login-body-card-form-divider { display: none !important; } .login-body-alternate-link { display: none !important; }" >> "$${css_file}"; \
		if ! grep -q "\.login-body-card-form { display: none !important; }" "$${css_file}"; then \
			echo "::error title=Hiding login form Verification Error::Failed to append rules to $${css_file}."; \
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

# ------------------------------------------------------------------------------
# Favicon replacement
# ------------------------------------------------------------------------------
# The PNGs under build/custom-assets/favicon/ replace the Mattermost favicons in
# the built webapp so that no volumeMount override is needed at deploy time.
#
# The webapp references favicons in two ways:
#   1. <link rel="icon"> in root.html -> dist/images/favicon/favicon-<variant>-<size>.png
#   2. unreads_status_handler.tsx imports the same PNGs, which webpack emits as
#      dist/files/<contenthash>.png and swaps into <link rel="icon"> at runtime
#      whenever the unread/mention state changes.
# The content hashes in (2) are computed from the upstream PNGs and change with
# the webpack toolchain, so they are resolved from the chunk containing the
# handler instead of being hard-coded. The handler imports the icons in the
# order default -> mentions -> unread, each in 16, 24, 32, 64, 96, and webpack
# emits the module constants in that order; the size of every target is
# verified before it is overwritten so a change in that order fails the build.
CUSTOM_ASSETS_DIR := $(dir $(lastword $(MAKEFILE_LIST)))custom-assets
CUSTOM_FAVICON_VARIANTS := default mentions unread
CUSTOM_FAVICON_SIZES := 16 24 32 64 96

customize-assets: customize-favicons

customize-favicons:
	@echo "replacing favicons..."
	@echo "CUSTOM_ASSETS_DIR = $(CUSTOM_ASSETS_DIR)"
	@for variant in $(CUSTOM_FAVICON_VARIANTS); do \
		for size in $(CUSTOM_FAVICON_SIZES); do \
			name="favicon-$${variant}-$${size}x$${size}.png"; \
			src="$(CUSTOM_ASSETS_DIR)/favicon/$${name}"; \
			dst="$(CUSTOMIZE_SOURCE_DIR)/images/favicon/$${name}"; \
			if [ ! -f "$${src}" ]; then \
				echo "::error title=Replacing favicon Error::$${src} not found."; \
				exit 1; \
			fi; \
			if [ ! -f "$${dst}" ]; then \
				echo "::error title=Replacing favicon Error::$${dst} not found. Upstream code might have changed."; \
				exit 1; \
			fi; \
			echo "-> $${dst} <- $${name}"; \
			cp "$${src}" "$${dst}"; \
		done; \
	done
	@handler_files=$$(grep -l 'link\[rel="icon"\]\[sizes="16x16"\]' $(CUSTOMIZE_SOURCE_DIR)/*.js 2>/dev/null); \
	if [ "$$(echo "$${handler_files}" | grep -c .)" -ne 1 ]; then \
		echo "::error title=Replacing favicon Error::Expected exactly one JS chunk containing the favicon handler, got: $${handler_files}. Upstream code might have changed."; \
		exit 1; \
	fi; \
	echo "-> Found handler chunk: $${handler_files}"; \
	hashed=$$(grep -oE 'files/[0-9a-f]+\.png' "$${handler_files}"); \
	if [ "$$(echo "$${hashed}" | grep -c .)" -ne 15 ]; then \
		echo "::error title=Replacing favicon Error::Expected 15 hashed favicon references in $${handler_files}, got: $${hashed}. Upstream code might have changed."; \
		exit 1; \
	fi; \
	i=0; \
	for variant in $(CUSTOM_FAVICON_VARIANTS); do \
		for size in $(CUSTOM_FAVICON_SIZES); do \
			i=$$((i + 1)); \
			name="favicon-$${variant}-$${size}x$${size}.png"; \
			src="$(CUSTOM_ASSETS_DIR)/favicon/$${name}"; \
			dst="$(CUSTOMIZE_SOURCE_DIR)/$$(echo "$${hashed}" | sed -n "$${i}p")"; \
			if [ ! -f "$${dst}" ]; then \
				echo "::error title=Replacing favicon Error::$${dst} not found."; \
				exit 1; \
			fi; \
			actual=$$(od -An -tu1 -j16 -N8 "$${dst}" | awk 'NR==1{printf "%dx%d", $$1*16777216+$$2*65536+$$3*256+$$4, $$5*16777216+$$6*65536+$$7*256+$$8}'); \
			if [ "$${actual}" != "$${size}x$${size}" ]; then \
				echo "::error title=Replacing favicon Verification Error::$${dst} is $${actual}, expected $${size}x$${size} for $${name}. Import order in unreads_status_handler might have changed."; \
				exit 1; \
			fi; \
			echo "-> $${dst} <- $${name}"; \
			cp "$${src}" "$${dst}"; \
			if ! cmp -s "$${src}" "$${dst}"; then \
				echo "::error title=Replacing favicon Verification Error::$${dst} differs from $${src} after copy."; \
				exit 1; \
			fi; \
		done; \
	done
	@echo "✅ Completed customize-favicons"
