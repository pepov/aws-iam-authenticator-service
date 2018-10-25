#!/bin/bash -e

: ${YQ_VER:="2.1.1"}
: ${EKSCTL_VER:="0.1.6"}
: ${REGION:=us-west-2}
: ${TOKEN_SERVICE:="http://localhost:8080"}
: ${INSTALL_LOCATION:=$PWD/bin}

if [[ " $@ " = *" -h "* ]]; then
    echo aws.sh creates an EKS cluster with a long live cluster-admin user authenticated with certification
    echo Common Options:
    echo -y auto confirm operations
    exit 0
fi

OS=$(uname -s | tr '[:upper:]' '[:lower:]')

echo -n "Enter the name of the cluster: "
read -t 30 CLUSTER_NAME

if [[ -z $AWS_ACCESS_KEY_ID ]]; then
    echo -n "Enter your AWS_ACCESS_KEY_ID: "
    read -s -t 30 AWS_ACCESS_KEY_ID
    echo
    export AWS_ACCESS_KEY_ID
fi
if [[ -z $AWS_SECRET_ACCESS_KEY ]]; then
    echo -n "Enter your AWS_SECRET_ACCESS_KEY: "
    read -s -t 30 AWS_SECRET_ACCESS_KEY
    echo
    export AWS_SECRET_ACCESS_KEY
fi

if ! which python > /dev/null; then
    echo "Please install python first."
    exit 128
fi

if [[ " $@ " = *" -y "* ]]; then
    AUTO_ACCEPT=true
fi

ask() {
    echo -n $1 "(y/N) "
    if [[ "$AUTO_ACCEPT" ]]; then
        echo y
    else
        read -t 600 ok
        if ! [[ '|Y|y|' = *"|$ok|"* ]]; then
            exit 1
        fi
    fi
}

is_linux() {
    [[ $OS == linux ]]
}

mkdir -p $INSTALL_LOCATION || :

if ! which pip > /dev/null; then
    ask "Installing pip"
    curl -sL https://bootstrap.pypa.io/get-pip.py | python
fi

USER_BASE=$(python -m site --user-base)

if ! [[ -f $USER_BASE/bin/aws ]]; then
    ask "Installing awscli"
    pip install --user awscli
fi

if ! [[ -f $INSTALL_LOCATION/yq ]]; then
    curl -sL https://github.com/mikefarah/yq/releases/download/${YQ_VER}/yq_${OS}_amd64 -o $INSTALL_LOCATION/yq
    chmod +x $INSTALL_LOCATION/yq
fi

if ! [[ -f $INSTALL_LOCATION/eksctl ]]; then
    curl -sL https://github.com/weaveworks/eksctl/releases/download/${EKSCTL_VER}/eksctl_$(uname -s)_amd64.tar.gz | tar xz -C $INSTALL_LOCATION
fi

export PATH=$INSTALL_LOCATION:$USER_BASE/bin:$PATH

KUBE_CONFIG_HEPTIO=config/$CLUSTER_NAME/.kube-config-heptio.yaml
KUBE_CONFIG=config/$CLUSTER_NAME/kube-config.yaml

if ! eksctl get cluster --name ${CLUSTER_NAME} >> /dev/null; then
    ask "Creating EKS cluster"
    mkdir -p $PWD/config/$CLUSTER_NAME || :
    eksctl create cluster --kubeconfig=${KUBE_CONFIG_HEPTIO} --name ${CLUSTER_NAME} --region ${REGION} --ssh-access
fi

cp -f ${KUBE_CONFIG_HEPTIO} ${KUBE_CONFIG}
yq w -i ${KUBE_CONFIG} users\[0\].user.exec.command curl
yq d -i ${KUBE_CONFIG} users\[0\].user.exec.args
yq w -i ${KUBE_CONFIG} users\[0\].user.exec.args\[+\] '"-s"'
yq w -i ${KUBE_CONFIG} users\[0\].user.exec.args\[+\] "${TOKEN_SERVICE}?clusterName=${CLUSTER_NAME}&awsAccessKeyID=${AWS_ACCESS_KEY_ID}&awsSecretAccessKey=${AWS_SECRET_ACCESS_KEY}"

echo Kubernetes config file is generated: ${PWD}/${KUBE_CONFIG}