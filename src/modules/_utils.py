import asyncio

async def has_audio_stream(url: str) -> bool:
    cmd = [
        'ffprobe',
        '-v', 'error',
        '-select_streams', 'a',
        '-show_entries', 'stream=index',
        '-of', 'csv=p=0',
        '-user_agent', 'Mozilla/5.0',
        url
    ]

    try:
        process = await asyncio.create_subprocess_exec(
            *cmd,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE
        )

        stdout, stderr = await asyncio.wait_for(process.communicate(), timeout=10)

        return bool(stdout.strip())
    except Exception as e:
        print(f"Error checking audio stream: {e}")
        return False
