# Device Status
The device status goes hand in hand with the rediscovery of the cameras, but goes beyond the scope of just discovery. 
It is a separate background task running at a specified interval (default 30s) to determine the most accurate 
operating status of the existing cameras. This applies to all devices regardless of how or where they were added from.

## States and Descriptions
Currently, there are 4 different statuses that a camera can have

- **UpWithAuth**: Can execute commands requiring credentials  
- **UpWithoutAuth**: Can only execute commands that do not require credentials. Usually this means the camera's credentials have not been registered with the service yet, or have been changed.  
- **Reachable**: Can be discovered but no commands can be received.  
- **Unreachable**: Cannot be seen by service at all. Typically, this means that there is a connection issue either physically or with the network.

### Status Check flow for each device
```mermaid
%% Note: The node and edge definitions are split up to make it easier to adjust the
%% links between the various nodes.
graph TD;
    %% -------- Node Definitions -------- %%
    update[Update Device Status<br/>in Core-Metadata]
    seen[Set `LastSeen` = Now]
    meta[Update Core-Metadata]
    with[UpWithAuth]
    without[UpWithoutAuth]
    reach[Reachable]
    unreach[Unreachable]
    check{Status Changed<br/>&&<br/>Status == UpWithAuth?}
    hasmac{Device Has<br/>MAC Address?}
    client[Create Onvif Client]
    caps[Device::GetCapabilities]
    macupdate[Check CredentialsMap for<br/>updated MAC Address]
    tcp[TCP Probe]
    info[GetDeviceInformation]
    uinfo[Update Device Information]
    umac[Update MAC Address]
    uref[Update EndpointRefAddress]
    unk{Device Name<br/>begins with<br/>unknown_unknown_?}
    remove[Remove Device<br/>unknown_unknown_]
    readd[Create Device<br/>&ltMfg&gt-&ltModel&gt-&ltEndpointRef&gt]
    
    %% -------- Graph Definitions -------- %%
    hasmac--No-->macupdate
    hasmac--Yes-->client
    macupdate-->client
    
    subgraph Test Connection Methods
        client-->caps
        caps--Failed-->tcp
        caps--Success-->info
        info--Success-->with
        info--Failed-->without
        tcp--Failed-->unreach
        tcp--Success-->reach
    end
    
    with-->seen
    without-->seen
    reach-->seen
    unreach-->update
    update-->check
    seen-->update
    check--Yes-->uinfo
    
    subgraph Refresh Device
        uinfo-->umac
        umac-->uref
        uref-->unk
        unk--No-->meta
        unk--Yes-->remove
        remove-->readd
    end
```

## Configuration Options
- Use `EnableStatusCheck` to enable the device status background service.
- `CheckStatusInterval` is the interval at which the service will determine the status of each camera.

```toml
EnableStatusCheck = true

# The interval in seconds at which the service will check the connection of all known cameras and update the device status 
# A longer interval will mean the service will detect changes in status less quickly
# Maximum 300s (1 hour)
CheckStatusInterval = 30
```

## Automatic Triggers
Currently, there are some actions that will trigger an automatic status check:
- Any modification to the `CredentialsMap` from the config provider (Consul)
