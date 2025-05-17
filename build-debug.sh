energy build --buildargs -tags=debug
mkdir -p debug
cp NordVPNGUI debug
cp -R ui debug
rm NordVPNGUI
cd energy
cp -r $(find . -type d -name 'CEF*')/* ../debug