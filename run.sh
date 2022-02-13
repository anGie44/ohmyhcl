#/usr/local/bin/bash

OUTFILE=import.sh
ERRORFILE=errors.log

echo "#/usr/local/bin/bash" >> $OUTFILE

# 1. Generate json of tfstate if not supplied

TFSTATE_JSON=$1
if [ -z "$TFSTATE_JSON" ]; then
  terraform show -json > temp.json
  TFSTATE_JSON=temp.json
fi

# 2. Run Migration
go run main.go -provider aws -provider-version "~> 4.0" -service s3 -input main.tf -output migrated.tf

#. 3. For each new resource, find the corresponding bucket ID in tfstate to generate import statements
while IFS=, read -r field1 field2
do
    ID=$(jq -r '.values[].resources[] | select(.address == "'"$field2"'") | .values.id' $TFSTATE_JSON)
    if [ -z "$ID" ]; then
      echo "[ERROR] unable to determine import ID for $field1" >> $ERRORFILE
    fi
    echo "terraform import $field1 $ID" >> $OUTFILE
done < ./output/resources.csv

