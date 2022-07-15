<!-- 
TODO: 
    Explanations of everything, not just images.
    Reword heading.
    Link to other docs.
    Mention to follow other guides first     
-->

# Camera Management Example App Service
The purpose of the camera management app is to automatically discover and connect to nearby ONVIF based cameras, 
control cameras via commands, create inference pipelines for the camera video streams and publish inference 
results to MQTT broker.

This app uses [EdgeX Core Services][edgex-core-services], [EdgeX ONVIF device service][device-onvif-camera] and [Edge Video Analytics Microservice][evam].

## Table of Contents

[Get the Source](#get-the-source)  
[Run Edge Analytics](#run-the-edge-video-analystics-microservice)  
[Configure and Build the App](#configure-and-build-the-app)  
[Using the App](#using-the-app)  
&nbsp;&nbsp;&nbsp;&nbsp;[Camera Position](#camera-position)  
&nbsp;&nbsp;&nbsp;&nbsp;[Start a Pipeline](#start-an-edge-video-analytics-pipeline)  
&nbsp;&nbsp;&nbsp;&nbsp;[Running Pipelines](#running-pipelines)  
&nbsp;&nbsp;&nbsp;&nbsp;[API Log](#api-log)  
[Next Steps](#next-steps)  


## Get the Source

1. Clone the repository.

    ```bash
    git clone https://github.com/edgexfoundry/edgex-examples.git
    ```

1. Navigate to the camera-management folder.

    ```bash
    cd application-services/custom/camera-management
    ```

## Run the Edge Video Analystics Microservice

1. Install Edge Video Analytics
    ```shell
    make install-edge-video-analytics
    ```

1. Run Edge Video Analytics in a dedicated terminal.
    ```bash
    make run-edge-video-analytics
    ```

## Configure and Build the App
1. Configure Camera Credentials

    1. Option 1: Modify the `configuration.toml` file and set the username and password to match that of the cameras.
        ```toml
        [Writable]
        ...
            [Writable.InsecureSecrets]
            ...
                [Writable.InsecureSecrets.CameraCredentials]
                path = "CameraCredentials"
                    [Writable.InsecureSecrets.CameraCredentials.Secrets]
                    username = ""
                    password = ""   
        ```
   
    1. Option 2: Export environment variable overrides
        ```shell
        export WRITABLE_INSECURESECRETS_CAMERACREDENTIALS_SECRETS_USERNAME=<username>
        export WRITABLE_INSECURESECRETS_CAMERACREDENTIALS_SECRETS_PASSWORD=<passowrd>
        ```

1. Build the app
    ```bash
    make build-app
    ```

1. Run the app
    ```bash
    make run-app
    ```
    >NOTE: By default the log will write in the terminal

# Using the App

1. Visit https://localhost:59750 to access the app.

![homepage](./images/homepage-demo-app-1.png)
    <p align="left">
        <i>Figure 1: Homepage for the Camera Management app</i>
    </p>

## Camera Position

You can control the position of supported cameras using ptz commands.  

![camera-position](./images/camera-position.png)

1. Use the arrows to control the direction of the camera movement. 
1. Use the magnifying glass icons to control the camera zoom.

## Start an Edge Video Analytics Pipeline

This section outlines how to start an analytics pipeline for inferencing on a specific camera stream.

![camera](./images/camera.png)

1. Select a camera out of the drop down list of connected cameras.  
    ![select-camera](./images/select-camera.png)

1. Select a video stream out of the drop down list of connected cameras.  
    ![select-profile](./images/select-profile.png)

1. Select a analytics pipeline out of the drop down list of connected cameras.  
    ![select-pipeline](./images/select-pipeline.png)

1. Click the `Start Pipeline` button.


## Running Pipelines

Once the pipeline is running, you can view the pipeline and its status.

![default pipelines state](./images/multiple-pipelines-default.png)  

1. Expand a pipeline to see its status. This includes important information aush as elapsed time, latency, frames per second, and elapsed time.
    ![select-camera](./images/running-pipelines.png)  

1. In the terminal where you started the app, once the pipeline is started, this log message will pop up. 
    ```bash
    level=INFO ts=2022-07-11T22:26:11.581149638Z app=app-camera-management source=evam.go:115 msg="View inference results at 'rtsp://<SYSTEM_IP_ADDRESS>:8554/<device name>'"
    ```

1. Use the URI from the log to view the camera footage with analytics overlayed.
    ```bash
    ffplay 'rtsp://<SYSTEM_IP_ADDRESS>:8554/<device name>'
    ```

    Example Output:
    ![example analytics](./images/example-analytics.png)

1. Press the red square stop button to shut down the pipeline.


## API Log

The API log shows the status of the 5 most recent calls and commands that the management has made. This includes important information from the responses, including camera information or error messages.

![API Logs](./images/api-log.png)  

1. Expand a log item to see the response  

    Good response: 
        ![good api response](./images/good-response.png) 
    Bad response: 
        ![bad api response](./images/bad-response.png)   

## Inference Events

![inference events default](./images/inference-events-default.png)   

1. To view the inference events in a json format, click the `Stream Events` button.

![inference events](./images/inference-events.png)  

## Next Steps

# License

[Apache-2.0](https://github.com/edgexfoundry-holding/device-onvif-camera/blob/main/LICENSE)


[edgex-core-services]: https://github.com/edgexfoundry/edgex-go
[device-onvif-camera]: https://github.com/edgexfoundry-holding/device-onvif-camera
[evam]: https://www.intel.com/content/www/us/en/developer/articles/technical/video-analytics-service.html


