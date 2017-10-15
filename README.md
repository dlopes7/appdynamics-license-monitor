## AppDynamics License Monitor - GO

This project uses the AppDynamics Api to get information about license usage and expiration date.

It can be used as an AppDynamics extension to generate metrics with the license usage.

It can monitor several different controllers, it uses goroutines to run the operations in parallel, giving very fast response times.

To use:
1. Download https://github.com/dlopes7/appdynamics-license-monitor/releases/download/1.0.0/license-monitor.zip
2. Unzip to the <machine_agent>/monitors folder
3. Edit config.json with your controller(s) 