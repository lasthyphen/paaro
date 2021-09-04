PKG_ROOT=/tmp/paaro
RPM_BASE_DIR=$PKG_ROOT/yum
DIJETS_BUILD_BIN_DIR=$RPM_BASE_DIR/usr/local/bin
DIJETS_LIB_DIR=$RPM_BASE_DIR/usr/local/lib/paaro

mkdir -p $RPM_BASE_DIR
mkdir -p $DIJETS_BUILD_BIN_DIR
mkdir -p $DIJETS_LIB_DIR

OK=`cp ./build/paaro $DIJETS_BUILD_BIN_DIR`
if [[ $OK -ne 0 ]]; then
  exit $OK;
fi
OK=`cp ./build/plugins/evm $DIJETS_LIB_DIR`
if [[ $OK -ne 0 ]]; then
  exit $OK;
fi

echo "Build rpm package..."
VER=$(echo $TAG | gawk -F- '{print$1}' | tr -d 'v' )
REL=$(echo $TAG | gawk -F- '{print$2}')
[ -z "$REL" ] && REL=0 
echo "Tag: $VER"
rpmbuild --bb --define "version $VER" --define "release $REL" --buildroot $RPM_BASE_DIR .github/workflows/yum/specfile/paaro.spec
aws s3 cp ~/rpmbuild/RPMS/x86_64/paaro-*.rpm s3://$BUCKET/linux/rpm/
