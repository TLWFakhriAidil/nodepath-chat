@echo off
echo Testing Campaign Trigger...
echo.

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src

echo Step 1: Building application...
set CGO_ENABLED=0
go build -o test_trigger.exe .

echo.
echo Step 2: Running campaign trigger test...
echo.

REM Create a simple Go test file
echo package main > test_trigger.go
echo. >> test_trigger.go
echo import ( >> test_trigger.go
echo     "fmt" >> test_trigger.go
echo     "github.com/aldinokemal/go-whatsapp-web-multidevice/config" >> test_trigger.go
echo     "github.com/aldinokemal/go-whatsapp-web-multidevice/database" >> test_trigger.go
echo     "github.com/aldinokemal/go-whatsapp-web-multidevice/usecase" >> test_trigger.go
echo     _ "github.com/lib/pq" >> test_trigger.go
echo ) >> test_trigger.go
echo. >> test_trigger.go
echo func main() { >> test_trigger.go
echo     config.Load() >> test_trigger.go
echo     database.InitDB() >> test_trigger.go
echo     fmt.Println("Testing campaign trigger...") >> test_trigger.go
echo     cts := usecase.NewCampaignTriggerService() >> test_trigger.go
echo     err := cts.ProcessCampaignTriggers() >> test_trigger.go
echo     if err != nil { >> test_trigger.go
echo         fmt.Printf("Error: %%v\n", err) >> test_trigger.go
echo     } else { >> test_trigger.go
echo         fmt.Println("Campaign trigger processed successfully!") >> test_trigger.go
echo     } >> test_trigger.go
echo } >> test_trigger.go

echo.
echo Step 3: Running the test...
go run test_trigger.go

echo.
echo Check your Worker Status page now!
pause