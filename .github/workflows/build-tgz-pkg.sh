PKG_ROOT=/tmp
VERSION=$TAG
DIJETS_ROOT=$PKG_ROOT/paaro-$VERSION

mkdir -p $DIJETS_ROOT

OK=`cp ./build/paaro $DIJETS_ROOT`
if [[ $OK -ne 0 ]]; then
  exit $OK;
fi
OK=`cp -r ./build/plugins $DIJETS_ROOT`
if [[ $OK -ne 0 ]]; then
  exit $OK;
fi


echo "Build tgz package..."
cd $PKG_ROOT
echo "Version: $VERSION"
tar -czvf "paaro-linux-$ARCH-$VERSION.tar.gz" paaro-$VERSION
aws s3 cp paaro-linux-$ARCH-$VERSION.tar.gz s3://$BUCKET/linux/binaries/ubuntu/$RELEASE/$ARCH/
