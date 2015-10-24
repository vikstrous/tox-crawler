#!/bin/bash
OUT=coverage.out
TMP=profile.out

echo "mode: atomic" > $OUT
for Dir in $(find ./* -maxdepth 10 -type d ); 
do
	if ls $Dir/*.go &> /dev/null;
	then
		returnval=`go test -race -coverprofile=profile.out $Dir`
		echo ${returnval}
		if [[ ${returnval} != *FAIL* ]]
		then
    		if [ -f $TMP ]
    		then
        		cat $TMP | grep -v "mode: atomic" >> $OUT
    		fi
    	else
    		exit 1
    	fi	
    fi
done

rm -rf $TMP
