git init               
git add .
git commit -m "second commit"
git branch -M main
git remote add origin git@github.com:IntelliSys-Lab/openwhisk-runtime-go-sui.git

sleep 1
git push -u origin main

