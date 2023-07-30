#!/usr/bin/env bash

# Copyright 2023 The KubeStellar Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

echo -e

echo "< Starting Kubestellar container >-------------------------"

# mkdir -p /kubestellar-logs
# chown -R $USER:$USER .kcp logs

# Start kcp
echo "< Starting kcp >-------------------------------------------"

echo -n "Running kcp... "
if [ -z "$EXTERNAL_HOSTNAME" ]; then
    kcp start &>> /kubestellar-logs/kcp.log &
else
    kcp start --external-hostname "$EXTERNAL_HOSTNAME" &>> /kubestellar-logs/kcp.log &
fi
echo "logfile=/kubestellar-logs/kcp.log"

echo "Waiting for kcp to be ready... it may take a while"
until [ "$(kubectl ws root:compute 2> /dev/null)" != "" ]
do
    sleep 1
done

echo "kcp version: $(kubectl version --short 2> /dev/null | grep kcp | sed 's/.*kcp-//')"

kubectl ws root


# Starting KubeStellar
echo "< Starting KubeStellar >-----------------------------------"

kubestellar start

# Sleep forerver
echo "Ready!"
sleep infinity
