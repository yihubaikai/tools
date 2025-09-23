import asyncio
import random
import subprocess
import datetime
from telethon import TelegramClient, functions
from telethon.errors import FloodWaitError

api_id = 28842458
api_hash = "c7a9b529ae626ea111b3ede0902ea0a3"
phone = "+22994495143"

# 排除名单（不发消息的人）
exclude_list = ["@chenhaonanaaaa", "@Dashuai522","@BotFather", "Telegram","@SpamBot", "@smss", "@MissRose_bot", "@41222", "@okpay", "@JiuguanYYBot"]

client = TelegramClient("session_22994495143", api_id, api_hash)

log_file = "log.txt"


def write_log(text: str):
    """写入日志文件"""
    with open(log_file, "a", encoding="utf-8") as f:
        f.write(f"[{datetime.datetime.now()}] {text}\n")


async def generate_message(username: str) -> str:
    """调用本地 ollama gemma3 模型生成一句推广话术"""
    prompt = f"写一句服务器销售推广介绍，少于30个字，包含问候，目标用户：{username}"
    result = subprocess.run(
        ["ollama", "run", "gemma3", prompt],
        capture_output=True
    )

    try:
        output = result.stdout.decode("utf-8", errors="ignore").strip()
    except Exception:
        output = ""

    if not output:
        output = f"你好 {username}，推荐高性能服务器，价格实惠！"

    return output


async def send_to_all():
    print("开始新一轮群发 ...")

    # 🚀 获取好友（兼容新版 Telethon）
    contacts_result = await client(functions.contacts.GetContactsRequest(hash=0))
    contacts = contacts_result.users

    # 🚀 获取所有私聊对话（排除群/频道）
    dialogs = []
    async for dialog in client.iter_dialogs():
        if dialog.is_user:  # 只要个人用户
            dialogs.append(dialog.entity)

    # 合并好友 + 私聊用户，去重
    users = {u.id: u for u in contacts + dialogs}.values()

    for entity in users:
        username = getattr(entity, "username", None)
        phone = getattr(entity, "phone", None)

        # 用户标识
        identifier = username or (phone and f"+{phone}") or getattr(entity, "first_name", "未知用户")

        # 排除名单检查
        if identifier in exclude_list or (username and f"@{username}" in exclude_list):
            print(f"跳过 {identifier}（在排除名单中）")
            write_log(f"跳过 {identifier}（在排除名单中）")
            continue

        try:
            message = await generate_message(identifier)
            await client.send_message(entity, message)
            print(f"发送给 {identifier}: {message}")
            write_log(f"成功发送给 {identifier}: {message}")

            # 随机延时，避免触发限流
            await asyncio.sleep(random.randint(5, 20))

        except FloodWaitError as e:
            print(f"触发限流，等待 {e.seconds} 秒后继续 ...")
            write_log(f"触发限流，等待 {e.seconds} 秒")
            await asyncio.sleep(e.seconds)
        except Exception as e:
            print(f"发送给 {identifier} 失败: {e}")
            write_log(f"发送给 {identifier} 失败: {e}")


async def main():
    while True:
        await send_to_all()
        print("本轮发送完成，等待 60 秒后再开始 ...")
        write_log("本轮发送完成，等待 60 秒后再开始 ...")
        await asyncio.sleep(60)


if __name__ == "__main__":
    with client:
        client.loop.run_until_complete(main())
