@echo off
title Windows RDP 端口修改工具
color 0a

echo.
echo ================================
echo   修改 Windows RDP 远程桌面端口
echo ================================
echo.

:: 输入新端口
set /p NEWPORT=请输入新的 RDP 端口（1025-65535，避免冲突）: 

if "%NEWPORT%"=="" (
    echo 端口不能为空！
    pause
    exit
)

:: 写入注册表
echo 正在修改注册表...
reg add "HKLM\SYSTEM\CurrentControlSet\Control\Terminal Server\WinStations\RDP-Tcp" /v PortNumber /t REG_DWORD /d %NEWPORT% /f

:: 修改防火墙规则
echo 添加防火墙规则...
netsh advfirewall firewall add rule name="RDP %NEWPORT%" dir=in action=allow protocol=TCP localport=%NEWPORT%

:: 尝试重启远程桌面服务（TermService）
echo 正在尝试重启 Remote Desktop Services 服务...
net stop TermService /y
net start TermService

echo.
echo ================================
echo RDP 端口已修改为 %NEWPORT%
echo 如果无法远程，请确认防火墙和安全组策略已放行。
echo ================================
pause
