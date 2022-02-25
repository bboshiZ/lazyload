go build 
cp lazyload manager
docker build -f docekrfile -t registry.ushareit.me/sgt/slime-lazyload:v1 .
docker push registry.ushareit.me/sgt/slime-lazyload:v1
