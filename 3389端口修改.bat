@echo off
title Windows RDP �˿��޸Ĺ���
color 0a

echo.
echo ================================
echo   �޸� Windows RDP Զ������˿�
echo ================================
echo.

:: �����¶˿�
set /p NEWPORT=�������µ� RDP �˿ڣ�1025-65535�������ͻ��: 

if "%NEWPORT%"=="" (
    echo �˿ڲ���Ϊ�գ�
    pause
    exit
)

:: д��ע���
echo �����޸�ע���...
reg add "HKLM\SYSTEM\CurrentControlSet\Control\Terminal Server\WinStations\RDP-Tcp" /v PortNumber /t REG_DWORD /d %NEWPORT% /f

:: �޸ķ���ǽ����
echo ��ӷ���ǽ����...
netsh advfirewall firewall add rule name="RDP %NEWPORT%" dir=in action=allow protocol=TCP localport=%NEWPORT%

:: ��������Զ���������TermService��
echo ���ڳ������� Remote Desktop Services ����...
net stop TermService /y
net start TermService

echo.
echo ================================
echo RDP �˿����޸�Ϊ %NEWPORT%
echo ����޷�Զ�̣���ȷ�Ϸ���ǽ�Ͱ�ȫ������ѷ��С�
echo ================================
pause
