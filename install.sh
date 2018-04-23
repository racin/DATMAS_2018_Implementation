#!/bin/bash
mkdir $HOME/.bcfs $HOME/.bcfs/Metadata $HOME/.bcfs/PubKeys $HOME/.bcfs/StorageSamples
cp configuration/test/clientConfig $HOME/.bcfs
cp configuration/test/appConfig $HOME/.bcfs
cp configuration/test/ipfsProxyConfig $HOME/.bcfs
cp configuration/test/accessControl_test $HOME/.bcfs/accessList
cd $HOME/.bcfs

# Generate Client RSA Keys
echo -ne '\n' | openssl req -x509 -nodes -newkey rsa:4096 -keyout client.pem -out client.pem
chmod 0600 client.pem
openssl rsa -in client.pem -pubout > client.pub
ssh-keygen -yf client.pem > client_ssh.pub
client_FP=$(ssh-keygen -E md5 -lf client_ssh.pub | egrep -o '([0-9a-f]{2}:){15}.{2}' | sed -E 's/://g')
rm client_ssh.pub
mv client.pub PubKeys/

# Generate Consensus RSA Keys
echo -ne '\n' | openssl req -x509 -nodes -newkey rsa:4096 -keyout consensus.pem -out consensus.pem
chmod 0600 consensus.pem
openssl rsa -in consensus.pem -pubout > consensus.pub
ssh-keygen -yf consensus.pem > consensus_ssh.pub
consensus_FP=$(ssh-keygen -E md5 -lf consensus_ssh.pub | egrep -o '([0-9a-f]{2}:){15}.{2}' | sed -E 's/://g')
rm consensus_ssh.pub
mv consensus.pub PubKeys/

# Generate Storage RSA Keys
echo -ne '\n' | openssl req -x509 -nodes -newkey rsa:4096 -keyout storage.pem -out storage.pem
chmod 0600 storage.pem
openssl rsa -in storage.pem -pubout > storage.pub
ssh-keygen -yf storage.pem > storage_ssh.pub
storage_FP=$(ssh-keygen -E md5 -lf storage_ssh.pub | egrep -o '([0-9a-f]{2}:){15}.{2}' | sed -E 's/://g')
rm storage_ssh.pub
mv storage.pub PubKeys/

sed -i -e 's/95c73e8028118d18a961dd1da6b5e7c3/'"$client_FP"'/g' accessList
sed -i -e 's/cc418e456ae72df5bdb39d65bb8945e8/'"$consensus_FP"'/g' accessList
sed -i -e 's/64168bb2f7a0a4d67d83471470ce757c/'"$storage_FP"'/g' accessList
sed -i -e 's/..\/crypto\/test_certificate\/client_test.pub/client.pub/g' accessList
sed -i -e 's/..\/crypto\/test_certificate\/storage_test.pub/storage.pub/g' accessList
sed -i -e 's/..\/crypto\/test_certificate\/consensus_test.pub/consensus.pub/g' accessList

sed -i -e 's/cc418e456ae72df5bdb39d65bb8945e8/'"$consensus_FP"'/g' appConfig
sed -i -e 's/cc418e456ae72df5bdb39d65bb8945e8/'"$consensus_FP"'/g' clientConfig
sed -i -e 's/64168bb2f7a0a4d67d83471470ce757c/'"$storage_FP"'/g' appConfig
sed -i -e 's/64168bb2f7a0a4d67d83471470ce757c/'"$storage_FP"'/g' clientConfig