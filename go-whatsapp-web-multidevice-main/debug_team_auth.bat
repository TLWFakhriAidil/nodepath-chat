@echo off
echo ========================================
echo Debugging Team Member Auth Issue
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

REM Add debug logging to the handler
echo Adding debug logging...

REM Create a debug version of the handler
echo // GetAllTeamMembers returns all team members with device counts (DEBUG VERSION) >> src\ui\rest\team_member_handlers_debug.txt
echo func (h *TeamMemberHandlers) GetAllTeamMembersDebug(c *fiber.Ctx) error { >> src\ui\rest\team_member_handlers_debug.txt
echo     // Log all context values >> src\ui\rest\team_member_handlers_debug.txt
echo     fmt.Printf("UserID from context: %%v\n", c.Locals("userId")) >> src\ui\rest\team_member_handlers_debug.txt
echo     fmt.Printf("IsTeamMember from context: %%v\n", c.Locals("isTeamMember")) >> src\ui\rest\team_member_handlers_debug.txt
echo     fmt.Printf("Request path: %%s\n", c.Path()) >> src\ui\rest\team_member_handlers_debug.txt
echo     fmt.Printf("Request method: %%s\n", c.Method()) >> src\ui\rest\team_member_handlers_debug.txt
echo     >> src\ui\rest\team_member_handlers_debug.txt
echo     // Try original logic >> src\ui\rest\team_member_handlers_debug.txt
echo     return h.GetAllTeamMembers(c) >> src\ui\rest\team_member_handlers_debug.txt
echo } >> src\ui\rest\team_member_handlers_debug.txt

echo.
echo Debug info created. The issue appears to be that the CustomAuth middleware
echo is not recognizing the admin user properly.
echo.
echo The probable cause is that the session/cookie authentication is not being
echo properly passed or recognized.
echo.
echo Let me check if there's a specific way admin authentication works...
pause
