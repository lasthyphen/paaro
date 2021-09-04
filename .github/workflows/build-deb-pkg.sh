PKG_ROOT=/tmp/paaro
DEBIAN_BASE_DIR=$PKG_ROOT/debian
DIJETS_BUILD_BIN_DIR=$DEBIAN_BASE_DIR/usr/local/bin
DIJETS_LIB_DIR=$DEBIAN_BASE_DIR/usr/local/lib/paaro
TEMPLATE=.github/workflows/debian/template 
DEBIAN_CONF=$DEBIAN_BASE_DIR/DEBIAN

mkdir -p $DEBIAN_BASE_DIR
mkdir -p $DEBIAN_CONF
mkdir -p $DIJETS_BUILD_BIN_DIR
mkdir -p $DIJETS_LIB_DIR

OK=`cp ./build/paaro $DIJETS_BUILD_BIN_DIR`
if [[ $OK -ne 0 ]]; then
  exit $OK;
fi
OK=`cp -r ./build/plugins $DIJETS_LIB_DIR`
if [[ $OK -ne 0 ]]; then
  exit $OK;
fi

OK=`cp $TEMPLATE/control $DEBIAN_CONF/control`
if [[ $OK -ne 0 ]]; then
  exit $OK;
fi

echo "Build debian package..."
cd $PKG_ROOT
echo "Tag: $TAG"
VER=$TAG
if [[ $TAG =~ ^v ]]; then
  VER=$(echo $TAG | tr -d 'v')
fi
NEW_VERSION_STRING="Version: $VER"
NEW_ARCH_STRING="Architecture: $ARCH"
sed -i "s/Version.*/$NEW_VERSION_STRING/g" debian/DEBIAN/control
sed -i "s/Architecture.*/$NEW_ARCH_STRING/g" debian/DEBIAN/control
dpkg-deb --build debian paaro-$TAG-$ARCH.deb
aws s3 cp paaro-$TAG-$ARCH.deb s3://${BUCKET}/linux/debs/ubuntu/$RELEASE/$ARCH/
rm -rf $PKG_ROOT
