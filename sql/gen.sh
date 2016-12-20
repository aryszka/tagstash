#! /bin/bash
set -e

sqldir=$(dirname $0)

goHead='package sql

// generated code
const Cmd_${name} = `
'

goTail='`'
 
genGo() {
	fn=$(basename $1)
	fnName=$(echo $fn | sed -e 's/[.].*//g')
	constName=$(echo $fnName | sed -e 's/-/_/g')
	echo "$goHead" | sed -e 's/[$]{name}/'"$constName"'/g' > $sqldir/$fnName.go
	cat $1 >> $sqldir/$fnName.go
	echo "$goTail" >> $sqldir/$fnName.go
}

for sql in $(ls $sqldir/*.sql); do genGo $sql; done
