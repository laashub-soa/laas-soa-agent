import asyncio
import time

import websockets


async def hello():
    async with websockets.connect('ws://localhost:8765') as websocket:
        while True:
            await websocket.send("hello world")
            greeting = await websocket.recv()
            print(f"< {greeting}")
            time.sleep(1)


def start_client():
    asyncio.get_event_loop().run_until_complete(hello())


try:
    start_client()
except Exception as e:
    print(e)
    time.sleep(2)
    start_client()
