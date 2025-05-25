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


credential = {"access_token":"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJkaWQiOiJkaWQ6cGxjOm1vcDdhaXF4M2RneGNpb3Z6bXg3bzZ4ZSIsImhhbmRsZSI6Im91eXphbTcucGRzLmF2YXRhci5haSIsImV4cGlyZWRfYXQiOiIwMDAxLTAxLTAxVDAwOjAwOjAwWiIsImlzcyI6ImF2YXRhcmFpLXNvY2lhbCIsInN1YiI6ImRpZDpwbGM6bW9wN2FpcXgzZGd4Y2lvdnpteDdvNnhlIiwiYXVkIjpbImF2YXRhcmFpLWFwcCJdLCJleHAiOjE3NDU1NTg4NDEsIm5iZiI6MTc0NTQ3MjQ0MSwiaWF0IjoxNzQ1NDcyNDQxLCJqdGkiOiI3ZTI3NzQ5MC0xODExLTRmYzYtYjFlOC1lODc3ZjZiZTg4ZDEifQ.KVhrjJsazHfc1jLY-alAVK9jjLijNotQN1pmUmaBKtI7JaX-OX01I0075shWa6pmzmPmLMX8qtJhznPbQYiEhYg45zyLUa1PRdxBOiy2NCVDEr8IQmuaf3-KM72l_Jauq1j23hwCm0zVomciOguupViNJHhiG-CP1B7VCO60KNp-tOLyiHC1dbfGsnRCsDWGNcRYx25lNkpe_AjrnHY6Q5ycISzsqVGxZQrkxtN_ZNpgED76WPRZsE280tgnbKXBYiESR8Ep1Q1Mppu6CRp8djCaiwFbqNJ5TRVZz8aTmYinDxni0MkkBfIXAubJHyk8rf5zOV7WdQap3mfzE_nLaA","did":"did:plc:mop7aiqx3dgxciovzmx7o6xe","expires_in":"86400","handle":"ouyzam7.pds.avatar.ai","refresh_token":"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJhdmF0YXJhaS1zb2NpYWwiLCJzdWIiOiI3ZTI3NzQ5MC0xODExLTRmYzYtYjFlOC1lODc3ZjZiZTg4ZDEiLCJhdWQiOlsiYXZhdGFyYWktYXBwIl0sImV4cCI6MTc0ODA2NDQ0MSwibmJmIjoxNzQ1NDcyNDQxLCJpYXQiOjE3NDU0NzI0NDEsImp0aSI6InJlZnJlc2gtN2UyNzc0OTAtMTgxMS00ZmM2LWIxZTgtZTg3N2Y2YmU4OGQxIn0.BOkofwSkwWUkVeHxZMk431U03AhGfTh2onjRtkyDe0cKjRiEA2OvZ3HPwifB_m2Lbvya-y8vsWYJSbHKU0BAP9uRj5Vdo6GU-EfAGf8DeMJshJ7apA8vLybNKiwN39vuVg-0RIV6mHgG_Xkookl7a4JlpgQ7-wkHnzU0rbEknsGty9Qw8kl7msCiumepIg7F68Cgukn7MdCJLVYd6wZ5SUjSaa37NQ9dGZYGBNKjKfqC00c7oLmD_53xfUsJjSoQY-nSXF1qeQ6IwkWHkQP5oHxiNVCAs-6D23UGE9O4fsFdNd7DosFfNY2jAgM1hT5Cjx-DlxJtldclPPbV_d62ww","token_type":"Bearer"}

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
    response = requests.get(url, headers=headers, verify=False, params={"refresh_token": credential["refresh_token"]})
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


if __name__ == "__main__":
    fire.Fire()
