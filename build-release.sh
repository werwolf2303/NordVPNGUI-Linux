energy build
mkdir -p release
cp NordVPNGUI release
cp -R ui release
rm NordVPNGUI
cd energy
cp -r $(find . -type d -name 'CEF*')/* ../release