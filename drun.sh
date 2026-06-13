# /bin/sh

PRJ_ROOT=$PRJ_ROOT
PRJ_NAME=$PRJ_NAME

# debug run
sudo docker run -it --rm \
    -v $PRJ_ROOT/etc:/app/etc \
    -v $PRJ_ROOT/var/log:/app/var/log \
    --net=host \
    $PRJ_NAME

