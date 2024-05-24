wget https://github.com/moby/buildkit/releases/download/v0.13.2/buildkit-v0.13.2.linux-amd64.tar.gz
tar -xzvf buildkit-v0.13.2.linux-amd64.tar.gz
rm buildkit-v0.13.2.linux-amd64.tar.gz
cp ./bin/* /usr/local/bin/
rm -r ./bin

cat <<EOF > /lib/systemd/system/buildkit.socket
[Unit]
Description=BuildKit
Documention=https://github.com/moby/buildkit

[Socket]
ListenStream=%t/buildkit/buildkitd.sock

[Install]
WantedBy=sockets.target
EOF

cat <<EOF > /lib/systemd/system/buildkitd.service
[Unit]
Description=BuildKit
Require=buildkit.socket
After=buildkit.socketDocumention=https://github.com/moby/buildkit

[Service]
ExecStart=/usr/local/bin/buildkitd --oci-worker=false --containerd-worker=true

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl start buildkitd
systemctl enable buildkitd