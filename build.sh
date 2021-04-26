# ssh-keygen -t rsa
# ssh-copy-id -i ~/.ssh/id_rsa.pub 192.168.1.107
ssh niuiic@192.168.1.104 "cd ~/AscendProjects/samples/cplusplus/level2_simple_inference/n_performance/1_multi_process_thread/face_recognition_camera && ./make.sh" > ./build.log 2>&1
sed -i "s/fatal/%ERROR% fatal/g" ./build.log
sed -i "s/\/home\/niuiic\/AscendProjects\/samples\/cplusplus\/level2_simple_inference\/n_performance\/1_multi_process_thread\/face_recognition_camera/\/home\/niuiic\/Documents\/Project\/Cpp\/face_recognition/g" ./build.log
sed -i "s/warning/%WARNING%/g" ./build.log
sed -i "s/error/%ERROR%/g" ./build.log
cat ./build.log
