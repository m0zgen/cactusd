#!/bin/bash
# Cretaed by Yevgeniy Gonvharov, https://lab.sys-adm.in

# Envs
# ---------------------------------------------------\
PATH=$PATH:/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin
SCRIPT_PATH=$(cd `dirname "${BASH_SOURCE[0]}"` && pwd); cd $SCRIPT_PATH; cd ..

DEST="/opt/cactusd/"

BUILD_PATH="$SCRIPT_PATH/builds"
BINARY_NAME="cactusd"

# Help information
usage() {

    echo -e "" "\nParameters:\n"
    echo -e "-b - Build Cactusd"
    echo -e "-d \"srv1 srv2 srv3 \" - Deploy Cactusd to targets. Can't works without -u\n"
    echo -e "-u - Remote user name"
    exit 1

}

# Date for backup file name
timestamp() {
    echo `date +%d-%m-%Y_%H-%M-%S`
}

checkingDirs() {

    echo "Checking $BUILD_PATH .."
    if [[ ! -d $BUILD_PATH/prev ]]; then
        mkdir -p $BUILD_PATH/prev
    fi
}

# Backup previous binary file
backupBinary() {

    if [[ -f "$BUILD_PATH/$BINARY_NAME" ]]; then
        bkp_name="cactusd-$(timestamp)"
        tar -zcvf $bkp_name.tar.gz $BUILD_PATH/$BINARY_NAME
        mv $bkp_name.tar.gz $BUILD_PATH/prev/
        echo "Previous release saved to $BUILD_PATH/prev/$bkp_name.tar.gz"
    fi
}

# Build current code to binary
buildBLD() {

    checkingDirs
    backupBinary

    echo "Building Cactusd release..."
    env GOOS=linux GOARCH=amd64 go build -o $BUILD_PATH
}


deployCactusd() {

    # buildBLD

    echo "Process deployment to server: $1 .."

    local cmdStop="sudo systemctl stop cactusd"
    local cmdStart="sudo systemctl start cactusd"

    ssh -ttt $1 "sudo systemctl stop cactusd"

    scp $SCRIPT_PATH/builds/cactusd $1:$DEST
    ssh -ttt $1 "sudo systemctl start cactusd"

}

if [[ -z "$1" ]]; then
    usage;
fi

# Checks arguments
while [[ "$#" -gt 0 ]]; do
    case $1 in
        -b|--build) BUILD=1; ;;
        -u|--user) USER=1; USERNAME=$2;;
        -d|--deploy) DEPLOY=1; TARGETS=$2; ;;
        -h|--help) usage ;; 
    esac
    shift
done

if [[ "$BUILD" -eq "1" ]]; then
    buildBLD; echo "Binary saved to: $SCRIPT_PATH/builds"; echo "Done!"
fi

if [[ "$DEPLOY" -eq "1" ]]; then
    if [[ -z "$TARGETS" ]]; then
        usage
    else
        if [[ -z "$USER" ]]; then
            usage
        else
            buildBLD
            for srv in $TARGETS; do
                deployCactusd $srv $USERNAME
            done
        fi
    fi
fi

# echo $DEPLOY $TARGETS


