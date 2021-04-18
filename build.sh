ssh niuiic@192.168.1.108 "cd ~/AscendProjects/samples/cplusplus/level2_simple_inference/n_performance/1_multi_process_thread/face_recognition_camera && ./make.sh" > ./build.log 2>&1
sed -i "s/\/home\/niuiic\/AscendProjects\/samples\/cplusplus\/level2_simple_inference\/n_performance\/1_multi_process_thread\/face_recognition_camera/\/home\/niuiic\/Documents\/Project\/Cpp\/face_recognition/g" ./build.log
sed -i "s/warning/_WARNING_/g" ./build.log
sed -i "s/error/_ERROR_/g" ./build.log
cat ./build.log
