import requests
import json
import base64
import os
from cryptography.hazmat.primitives.asymmetric import ed25519
from cryptography.hazmat.primitives import serialization

def create_did_with_plc():
    # 生成 Ed25519 密钥对
    private_key = ed25519.Ed25519PrivateKey.generate()
    public_key = private_key.public_key()

    # 将公钥转换为多格式编码
    public_key_bytes = public_key.public_bytes(
        encoding=serialization.Encoding.Raw,
        format=serialization.PublicFormat.Raw
    )
    # 使用正确的多格式编码方式
    public_key_multibase = f"did:key:z{base64.urlsafe_b64encode(public_key_bytes).decode('utf-8').rstrip('=')}"

    # 创建正确格式的操作文档
    operation = {
        "type": "create",
        "rotationKeys": [public_key_multibase],
        "verificationMethods": {
            "atproto": public_key_multibase
        },
        "handle": "username.yourdomain.com",  # 使用您可以验证的域名
        "service": {
            "atproto_pds": {
                "type": "AtprotoPersonalDataServer",
                "endpoint": "https://your-pds-endpoint.com"
            }
        }
    }

    # 序列化操作文档
    operation_bytes = json.dumps(operation, separators=(',', ':')).encode()

    # 使用私钥签名
    signature = private_key.sign(operation_bytes)
    signature_b64 = base64.urlsafe_b64encode(signature).decode('utf-8').rstrip('=')

    # 准备请求数据 - 注意格式
    request_data = {
        "operation": operation,
        "signature": signature_b64,
        "did": ""  # 新创建的 DID 将由服务器生成
    }

    # 发送到 PLC 服务的正确端点
    response = requests.post(
        "https://plc.avatar.ai/operation",  # 注意这里使用 /operation 而不是根路径
        json=request_data,
        headers={"Content-Type": "application/json"},
        verify=False
    )

    if response.status_code == 200:
        result = response.json()
        print(f"成功创建 DID: {result.get('did')}")
        return result
    else:
        print(f"创建 DID 失败: {response.status_code}")
        print(response.text)
        return None


if __name__ == "__main__":
    create_did_with_plc()