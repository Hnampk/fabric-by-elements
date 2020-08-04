if [ -z "$1" ]
  then
    echo "No chaincode name supplied!"
    exit 1
fi

if [ -f "code.tar.gz" ]; then
    rm code.tar.gz
fi

if [ -f *-pkg.tgz ]; then
    rm *-pkg.tgz
fi

tar cfz code.tar.gz connection.json src/

tar cfz $1-pkg.tgz metadata.json code.tar.gz