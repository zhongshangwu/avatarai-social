import fire
import requests
import base64
import mimetypes
token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJkaWQiOiJkaWQ6cGxjOm1vcDdhaXF4M2RneGNpb3Z6bXg3bzZ4ZSIsImhhbmRsZSI6Im91eXphbTcucGRzLmF2YXRhci5haSIsImV4cGlyZWRfYXQiOiIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsImlzcyI6ImF2YXRhcmFpLXNvY2lhbCIsInN1YiI6ImRpZDpwbGM6bW9wN2FpcXgzZGd4Y2lvdnpteDdvNnhlIiwiYXVkIjpbImF2YXRhcmFpLWFwcCJdLCJleHAiOjE3NDUxNjE4MzgsIm5iZiI6MTc0NTA3NTQzOCwiaWF0IjoxNzQ1MDc1NDM4LCJqdGkiOiIyNzQ1Y2I0Ni1iMTA2LTRhNDYtYjExNi00NDMwYTE3ODlhMjkifQ.PWaH6ugwwLKP0YsPE1KpRj25XDPl3VbNdgoYxAAfPzkcbK2zzcA3tQkgRlsqLj3Aa9cSNcptQUAQPWAVYQQIfnnx_hbBWzNxfyTVLPABgOMiMfICZrDBDTx-G4gSjkRKDZDSZn5VQfHC90rhzA_qP2DlIeiQxqOwOnFfLNfca_71fWPY26pzRClJB-ZqYMGQ6rKabXWIL3EVMnJHQIXQ1bWCA4pf4KiRpfjDC2WaA6XH6c_4y6yQuicCOWgerCu44qDsVR2xQ_Yv7O8aX2llEdryD3cq2qaoH1tZxz-vFcRhF39GaQc6r7TZs-1Mmb0uZNwEnkmVadjZR-25c0bp6g"
# curl 'https://avatarai.social/api/aster/profile' \
#   -H 'Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7' \
#   -H 'Accept-Language: zh-CN,zh;q=0.9,en;q=0.8,zh-TW;q=0.7' \
#   -H 'Cache-Control: max-age=0' \
#   -H 'Connection: keep-alive' \
#   -b 'avatarai_token=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJkaWQiOiJkaWQ6cGxjOm1vcDdhaXF4M2RneGNpb3Z6bXg3bzZ4ZSIsImhhbmRsZSI6Im91eXphbTcucGRzLmF2YXRhci5haSIsImV4cGlyZWRfYXQiOiIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsImlzcyI6ImF2YXRhcmFpLXNvY2lhbCIsInN1YiI6ImRpZDpwbGM6bW9wN2FpcXgzZGd4Y2lvdnpteDdvNnhlIiwiYXVkIjpbImF2YXRhcmFpLWFwcCJdLCJleHAiOjE3NDUxNjE4MzgsIm5iZiI6MTc0NTA3NTQzOCwiaWF0IjoxNzQ1MDc1NDM4LCJqdGkiOiIyNzQ1Y2I0Ni1iMTA2LTRhNDYtYjExNi00NDMwYTE3ODlhMjkifQ.PWaH6ugwwLKP0YsPE1KpRj25XDPl3VbNdgoYxAAfPzkcbK2zzcA3tQkgRlsqLj3Aa9cSNcptQUAQPWAVYQQIfnnx_hbBWzNxfyTVLPABgOMiMfICZrDBDTx-G4gSjkRKDZDSZn5VQfHC90rhzA_qP2DlIeiQxqOwOnFfLNfca_71fWPY26pzRClJB-ZqYMGQ6rKabXWIL3EVMnJHQIXQ1bWCA4pf4KiRpfjDC2WaA6XH6c_4y6yQuicCOWgerCu44qDsVR2xQ_Yv7O8aX2llEdryD3cq2qaoH1tZxz-vFcRhF39GaQc6r7TZs-1Mmb0uZNwEnkmVadjZR-25c0bp6g' \
#   -H 'Sec-Fetch-Dest: document' \
#   -H 'Sec-Fetch-Mode: navigate' \
#   -H 'Sec-Fetch-Site: none' \
#   -H 'Sec-Fetch-User: ?1' \
#   -H 'Upgrade-Insecure-Requests: 1' \
#   -H 'User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36' \
#   -H 'sec-ch-ua: "Google Chrome";v="135", "Not-A.Brand";v="8", "Chromium";v="135"' \
#   -H 'sec-ch-ua-mobile: ?0' \
#   -H 'sec-ch-ua-platform: "Windows"'


credential = {"access_token":
    "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJkaWQiOiJkaWQ6cGxjOm1vcDdhaXF4M2RneGNpb3Z6bXg3bzZ4ZSIsImhhbmRsZSI6Im91eXphbTcucGRzLmF2YXRhci5haSIsImV4cGlyZWRfYXQiOiIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsImlzcyI6ImF2YXRhcmFpLXNvY2lhbCIsInN1YiI6ImRpZDpwbGM6bW9wN2FpcXgzZGd4Y2lvdnpteDdvNnhlIiwiYXVkIjpbImF2YXRhcmFpLWFwcCJdLCJleHAiOjE3NDk2MzUwNTIsIm5iZiI6MTc0OTU0ODY1MiwiaWF0IjoxNzQ5NTQ4NjUyLCJqdGkiOiJmMTE2NzVlYi00ZmE4LTRkOWQtYmE5Zi00ZGM1NDhiOTliMzMifQ.niUbz-lcRWbGuKl6cBOuKf6jeJSoZV82CcIt2Fsyp3AebLjfsOKqMy2CoTM7E6PvlL306ulTBZ47KxI2rGE0W8jKLhZykhmhwd6XiDxdFvcVF_HMxlyHa1g3jQJeci3p3Fv7yrbmCTmnBnM2UwJJYNkpTux_Hq43ONfThOdabItFV60ITRd35feswJaj56oRNMPBwQXLut82vfbthuakhhWVfnasXKP9cNkA4CaciG5iIsmGARD4Q6FAzxePJiJTIEaxtLowjUfC36tmZ5Az9XygHH_2rdjAjh4og8lcdNNUvh-qgXTaZcaOLOM20Og9mFjTTiRTgnNWpEUAbSNN4Q",
    "did":"did:plc:mop7aiqx3dgxciovzmx7o6xe","expires_in":"86400",
    "handle":"ouyzam7.pds.avatar.ai",
    "refresh_token":"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJhdmF0YXJhaS1zb2NpYWwiLCJzdWIiOiJmMTE2NzVlYi00ZmE4LTRkOWQtYmE5Zi00ZGM1NDhiOTliMzMiLCJhdWQiOlsiYXZhdGFyYWktYXBwIl0sImV4cCI6MTc1MjE0MDY1MiwibmJmIjoxNzQ5NTQ4NjUyLCJpYXQiOjE3NDk1NDg2NTIsImp0aSI6InJlZnJlc2gtZjExNjc1ZWItNGZhOC00ZDlkLWJhOWYtNGRjNTQ4Yjk5YjMzIn0.D4LdgvIkHwFi1d_zNkjs2Mr8oefkjNJTvCHTM9y5Ouzb6M5VfSaiAFW8qtRfJdX0mQ5nETiO8tLOyFaSBHuUgkcm8MFU6zS1ftUEFspnfzfKoGscu30F_HUe3aeq9eo6kiHJB2Tvbq1M8CtQ2ZLtS9UfWyIZXTtZeXQeqRw3mDrrLyVeAg3_xaKPX-ZmSwcDVohGM5UUwf0inmM_Of0O7jx0b5dGJR976RfREU_xNPtDERxKDNhhzSJPL7Hlh6DqylIB_8UzjpKImTD6HGz0YvMY3DRWhHtNSGn6BVy7ALT5dPH-_uIsOGQRbqaNsLg9ev9ogS4bzqkJ5wTFeGBJ3A",
    "token_type":"Bearer"}

HOST = "http://127.0.0.1:8082"


def get_headers():
    token = credential["access_token"]
    return {
        "Authorization": f"Bearer {token}",
        # "Cookie": f"avatarai_token={token}",
        "Content-Type": "application/json",
    }

def get_auth_headers():
    token = credential["access_token"]
    return {
        "Authorization": f"Bearer {token}",
    }

def refresh_token():
    url = f"{HOST}/api/oauth/refresh"
    headers = get_headers()
    response = requests.post(url, headers=headers, verify=False, json={"refresh_token": credential["refresh_token"]})
    print(response.status_code)
    print(response.text)


def aster_profile():
    # https://avatarai.social/api/aster/profile
    url = f"{HOST}/api/aster/profile"
    headers = get_headers()
    response = requests.get(url, headers=headers, verify=False)
    print(response.status_code)
    print(response.text)

def aster_mint():
    # https://avatarai.social/api/aster/mint
    url = f"{HOST}/api/aster/mint"
    headers = get_headers()
    response = requests.post(url, headers=headers, verify=False)
    print(response.json())

def avatar_profile():
    # https://avatarai.social/api/avatar/profile
    url = f"{HOST}/api/avatar/profile"
    headers = get_headers()
    response = requests.get(url, headers=headers, verify=False)
    print(response.status_code)
    print(response.text)


def generate_fake_avatar(prefix="A"):
    # 创建一个数字图片
    # 使用 PIL 库生成一个数字图片
    from PIL import Image, ImageDraw, ImageFont
    import random
    import string

    # 创建一个数字图片
    image = Image.new('RGB', (100, 100), color = (255,255,255))
    draw = ImageDraw.Draw(image)
    font = ImageFont.load_default()

    number = random.randint(1000, 9999)

    draw.text((10,10), f"{prefix}{number}", fill=(0,0,0), font=font)
    image.save("test.png")

    # 将图片转换为 base64
    with open("test.png", "rb") as image_file:
        image_data = image_file.read()
        print(len(image_data), "size")
        return base64.b64encode(image_data).decode('utf-8')

def avatar_update_profile():
    # https://avatarai.social/api/avatar/profile
    url = f"{HOST}/api/avatar/profile"
    headers = get_headers()

    # 创建一个数字图片
    data = {
        "displayName": "test-display-name",
        "description": "test-description",
        "avatarBase64": generate_fake_avatar("A"),
        "bannerBase64": generate_fake_avatar("B"),
    }
    response = requests.post(url, headers=headers, json=data, verify=False)
    print(response.status_code)
    print(response.text)

def upload_blob():
    # https://avatarai.social/api/blobs
    url = f"{HOST}/api/blobs"
    generate_fake_avatar(prefix="C")
    headers = get_auth_headers()
    with open("test.png", "rb") as image_file:
        image_data = image_file.read()
        print(len(image_data), "size")
        # auto detect mime type
        mime_type = mimetypes.guess_type("test.png")[0]
        files = {'file': ('your_image.png', image_data, mime_type)}
        response = requests.post(url, headers=headers, files=files, verify=False)
    print(response.request.headers)
    print(response.status_code)
    print(response.text)

    return response.json()["blob"]


def create_moment():
    # https://avatarai.social/api/moments
    url = f"{HOST}/api/moments"
    headers = get_headers()
    from atproto import client_utils
    text_builder = client_utils.TextBuilder()
    text_builder.tag('This is a rich message. ', 'atproto')
    text_builder.text('I can mention ')
    text_builder.mention('account', 'did:plc:kvwvcn5iqfooopmyzvb4qzba')
    text_builder.text(' and add clickable ')
    text_builder.link('link', 'https://atproto.blue/')

    import json

    image_blob = upload_blob()

    data = {
        "text": text_builder.build_text(),
        "facets": [facet.model_dump(by_alias=True) for facet in text_builder.build_facets()],
        "images": [image_blob],
        "langs": ["en", "zh"],
        "tags": ["test"],
    }
    print(data)

    # data = {'text': 'This is a rich message. I can mention account and add clickable link', 'facets': [{'features': [{'tag': 'atproto', 'py_type': 'app.bsky.richtext.facet#tag'}], 'index': {'byte_end': 24, 'byte_start': 0, 'py_type': 'app.bsky.richtext.facet#byteSlice'}, 'py_type': 'app.bsky.richtext.facet'}, {'features': [{'did': 'did:plc:kvwvcn5iqfooopmyzvb4qzba', 'py_type': 'app.bsky.richtext.facet#mention'}], 'index': {'byte_end': 45, 'byte_start': 38, 'py_type': 'app.bsky.richtext.facet#byteSlice'}, 'py_type': 'app.bsky.richtext.facet'}, {'features': [{'uri': 'https://atproto.blue/', 'py_type': 'app.bsky.richtext.facet#link'}], 'index': {'byte_end': 68, 'byte_start': 64, 'py_type': 'app.bsky.richtext.facet#byteSlice'}, 'py_type': 'app.bsky.richtext.facet'}], 'images': [{'$type': 'blob', 'ref': {'$link': 'bafkreiajdw4zpt76hi4xt3ifcmfqmvlf66vd2vuiqsxtu7ulyayhj2y5jq'}, 'mimeType': 'image/png', 'size': 798}], 'langs': ['en'], 'tags': ['test']}
    # return
    response = requests.post(url, headers=headers, json=data, verify=False)
    print(response.status_code)
    print(response.text)


def moment_detail():
    # https://avatarai.social/api/moments/detail
    url = f"{HOST}/api/moments/detail"
    headers = get_headers()
    response = requests.get(url, headers=headers, verify=False, params={"uri": "at://did:plc:mop7aiqx3dgxciovzmx7o6xe/app.vtri.activity.moment/3lnjy7ahype22"})
    print(response.status_code)
    print(response.text)


def feed():
    # https://avatarai.social/api/feed
    url = f"{HOST}/api/feed"
    headers = get_headers()
    response = requests.get(url, headers=headers, verify=False)
    print(response.status_code)
    print(response.text)


def messages_history(room_id=None, thread_id=None, before=None, after=None, limit=None):
    """
    测试获取历史消息接口

    Args:
        room_id: 房间ID (必需)
        thread_id: 线程ID (必需)
        before: 获取指定消息ID之前的消息
        limit: 获取before/after消息的数量
        after: 获取指定消息ID之后的消息

    Examples:
        # 获取最新的20条消息
        python ava.py messages_history --room_id="room123" --thread_id="thread456"

        # 获取指定消息之前的10条消息
        python ava.py messages_history --room_id="room123" --thread_id="thread456" --before="msg_id" --limit=10

        # 获取指定消息之后的5条消息
        python ava.py messages_history --room_id="room123" --thread_id="thread456" --after="msg_id" --limit=5
    """
    # 默认测试数据
    if not room_id:
        room_id = "123"
    if not thread_id:
        thread_id = "123"

    url = f"{HOST}/api/messages/history"
    headers = get_headers()

    # 构建查询参数
    params = {
        "roomId": room_id,
        "threadId": thread_id
    }

    if before:
        params["before"] = before
    if limit:
        params["limit"] = limit
    if after:
        params["after"] = after

    print(f"请求URL: {url}")
    print(f"查询参数: {params}")

    response = requests.get(url, headers=headers, params=params, verify=False)

    print(f"响应状态码: {response.status_code}")
    print(f"响应内容: {response.text}")

    if response.status_code == 200:
        data = response.json()
        messages = data.get("messages", [])
        pagination = data.get("pagination", {})

        print(f"\n=== 消息历史查询结果 ===")
        print(f"消息数量: {len(messages)}")
        print(f"分页信息: {pagination}")

        if messages:
            print(f"\n=== 消息列表 ===")
            for i, msg in enumerate(messages):
                print(f"消息 {i+1}:")
                print(f"  ID: {msg.get('id')}")
                print(f"  类型: {msg.get('msgType')}")
                print(f"  发送者: {msg.get('senderId')}")
                print(f"  接收者: {msg.get('receiverId')}")
                print(f"  创建时间: {msg.get('createdAt')}")
                print(f"  内容: {msg.get('content')}")
                print()
        else:
            print("没有找到消息")
    else:
        print(f"请求失败: {response.text}")

    return response


def test_messages_history_scenarios():
    """
    测试多种历史消息查询场景
    """
    print("=== 开始测试历史消息接口的各种场景 ===\n")

    # 场景1: 获取最新消息（默认）
    print("场景1: 获取最新的消息")
    messages_history()
    print("\n" + "="*50 + "\n")

    # 场景2: 获取指定数量的最新消息
    print("场景2: 获取最新的5条消息")
    messages_history(limit=5)
    print("\n" + "="*50 + "\n")

    # 场景3: 获取指定消息之前的消息（需要先有消息ID）
    print("场景3: 获取指定消息之前的消息")
    print("注意: 这需要一个真实的消息ID，这里使用示例ID")
    messages_history(before="example_message_id", limit=3)
    print("\n" + "="*50 + "\n")

    # 场景4: 获取指定消息之后的消息
    print("场景4: 获取指定消息之后的消息")
    print("注意: 这需要一个真实的消息ID，这里使用示例ID")
    messages_history(after="example_message_id", limit=3)
    print("\n" + "="*50 + "\n")

    # 场景5: 测试错误情况 - 缺少必需参数
    print("场景5: 测试错误情况 - 缺少roomId参数")
    url = f"{HOST}/api/messages/history"
    headers = get_headers()
    params = {"threadId": "test_thread"}
    response = requests.get(url, headers=headers, params=params, verify=False)
    print(f"响应状态码: {response.status_code}")
    print(f"错误信息: {response.text}")
    print("\n" + "="*50 + "\n")

    print("=== 历史消息接口测试完成 ===")


if __name__ == "__main__":
    fire.Fire()
