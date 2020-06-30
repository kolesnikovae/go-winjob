# Job object demo

The program starts a process in a job object and waits for interruption, printing job notifications and counters. The following limits applied:
 * BreakawayOK
 * Kill on job close
 * Active process limit
 * CPU hard cap 5%
 * CPU process time (user) 10s
 * Process memory 10MB
 * Write clipboard

The program accepts path to the executable which is supposed to be run in a job object and its arguments, e.g.:
```
PS > .\job.exe cmd.exe /c "ping localhost -t"
``` 
If no arguments provided, *notepad.exe* will be started.

<details>
  <summary>Sample output</summary>  
 
  ```
  PS > .\job.exe cmd.exe /c "notepad.exe"
  2020/06/28 12:35:35 Notification: winjob.Notification{Type:"NewProcess", PID:5116}
  2020/06/28 12:35:35 Notification: winjob.Notification{Type:"NewProcess", PID:2088}
  2020/06/28 12:35:39 Counters:
  {
          "TotalUserTime": 0,
          "TotalKernelTime": 0,
          "ThisPeriodTotalUserTime": 0,
          "ThisPeriodTotalKernelTime": 0,
          "TotalPageFaultCount": 5955,
          "TotalProcesses": 2,
          "ActiveProcesses": 2,
          "TotalTerminatedProcesses": 0,
          "ReadOperationCount": 52,
          "WriteOperationCount": 0,
          "OtherOperationCount": 553,
          "ReadTransferCount": 202300,
          "WriteTransferCount": 0,
          "OtherTransferCount": 553
  }
  
  < omited >
  
  2020/06/28 12:35:53 Counters:
  {
          "TotalUserTime": 15625000,
          "TotalKernelTime": 15625000,
          "ThisPeriodTotalUserTime": 15625000,
          "ThisPeriodTotalKernelTime": 15625000,
          "TotalPageFaultCount": 7648,
          "TotalProcesses": 2,
          "ActiveProcesses": 2,
          "TotalTerminatedProcesses": 0,
          "ReadOperationCount": 52,
          "WriteOperationCount": 0,
          "OtherOperationCount": 619,
          "ReadTransferCount": 202300,
          "WriteTransferCount": 0,
          "OtherTransferCount": 619
  }
  2020/06/28 12:35:56 Notification: winjob.Notification{Type:"ExitProcess", PID:2088}
  2020/06/28 12:35:56 Notification: winjob.Notification{Type:"ExitProcess", PID:5116}
  2020/06/28 12:35:56 Notification: winjob.Notification{Type:"ActiveProcessZero", PID:0}
  2020/06/28 12:35:57 Counters:
  {
          "TotalUserTime": 15625000,
          "TotalKernelTime": 15625000,
          "ThisPeriodTotalUserTime": 15625000,
          "ThisPeriodTotalKernelTime": 15625000,
          "TotalPageFaultCount": 7900,
          "TotalProcesses": 2,
          "ActiveProcesses": 0,
          "TotalTerminatedProcesses": 0,
          "ReadOperationCount": 52,
          "WriteOperationCount": 0,
          "OtherOperationCount": 638,
          "ReadTransferCount": 202300,
          "WriteTransferCount": 0,
          "OtherTransferCount": 638
  }
  2020/06/28 12:36:01 Closing job object
  2020/06/28 12:36:01 Closing subscription
  ```

</details>
