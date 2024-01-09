export TEMPLATE_FILE="./submodules/api-openapi-spec/templates/website.yaml"
export WEBSITE_DOCS_DIRECTORY="./wp-content/themes/wikimedia-enterprise-theme/assets/docs/"

# Detect the platform
PLATFORM=$(uname)

# Set the appropriate yq binary based on the platform
if [ "$PLATFORM" = "Darwin" ]; then
  echo "Mac OS platform detected"
  YQ_BINARY="./submodules/api-openapi-spec/bin/yq"
else
  echo "Assuming linux platform"
  YQ_BINARY="./submodules/api-openapi-spec/bin/yq_linux_amd64"
fi

chmod +x $YQ_BINARY

echo "Ondemand JSON file generating..."
export ONDEMAND_OUTPUT_FILE="$WEBSITE_DOCS_DIRECTORY/ondemand.json";
rm -f $ONDEMAND_OUTPUT_FILE && $YQ_BINARY 'delpaths([["paths","/v2/snapshots*"], ["paths","/v2/batches*"],["paths", "/v2/codes*"], ["paths","/v2/namespaces*"],["paths","/v2/projects*"],["paths","/v2/languages*"]])' ./submodules/api-openapi-spec/main.yaml -o json | $YQ_BINARY '. * load("'$TEMPLATE_FILE'")' -o=json - >> $ONDEMAND_OUTPUT_FILE

echo "Snapshots JSON file generating..."
export SNAPSHOTS_OUTPUT_FILE="$WEBSITE_DOCS_DIRECTORY/snapshots.json";
rm -f $SNAPSHOTS_OUTPUT_FILE && $YQ_BINARY 'delpaths([["paths","/v2/batches*"],["paths","/v2/articles*"],["paths","/v2/structured-contents*"],["paths","/v2/codes*"],["paths","/v2/namespaces*"],["paths","/v2/projects*"],["paths","/v2/languages*"]])' ./submodules/api-openapi-spec/main.yaml -o json | $YQ_BINARY '. * load("'$TEMPLATE_FILE'")' -o=json - >> $SNAPSHOTS_OUTPUT_FILE

echo "Batches JSON file generating..."
export BATCHES_OUTPUT_FILE="$WEBSITE_DOCS_DIRECTORY/batches.json";
rm -f $BATCHES_OUTPUT_FILE && $YQ_BINARY 'delpaths([["paths","/v2/snapshots*"],["paths","/v2/articles*"],["paths","/v2/structured-contents*"],["paths", "/v2/codes*"],["paths","/v2/namespaces*"],["paths","/v2/projects*"],["paths","/v2/languages*"]])' ./submodules/api-openapi-spec/main.yaml -o json | $YQ_BINARY '. * load("'$TEMPLATE_FILE'")' -o=json - >> $BATCHES_OUTPUT_FILE

echo "General endpoints file generating..."
export GENERAL_OUTPUT_FILE="$WEBSITE_DOCS_DIRECTORY/general.json"
rm -f $GENERAL_OUTPUT_FILE && $YQ_BINARY 'delpaths([["paths","/v2/snapshots*"],["paths","/v2/articles*"],["paths","/v2/structured-contents*"],["paths","/v2/batches*"]])' ./submodules/api-openapi-spec/main.yaml -o json | $YQ_BINARY '. * load("'$TEMPLATE_FILE'")' -o=json - >> $GENERAL_OUTPUT_FILE

echo "Realtime JSON file generating..."
export REALTIME_OUTPUT_FILE="$WEBSITE_DOCS_DIRECTORY/realtime.json"
rm -f $REALTIME_OUTPUT_FILE && $YQ_BINARY ./submodules/api-openapi-spec/realtime.yaml -o json >> $REALTIME_OUTPUT_FILE
