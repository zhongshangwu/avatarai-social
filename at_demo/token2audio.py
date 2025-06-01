import asyncio
import threading
import time
from typing import List, Callable, Optional, Dict, Any
from dataclasses import dataclass
from queue import Queue, Empty
from concurrent.futures import ThreadPoolExecutor
import uuid

@dataclass
class AudioFrame:
  """音频帧数据结构"""
  data: bytes
  sample_rate: int
  channels: int

@dataclass
class Config:
  """配置结构"""
  endpoint: str = ""
  model: str = ""
  direct: str = ""
  token_type: str = "1o"  # 1o or 1f

class Worker:
  """工作线程类"""
  def __init__(self):
      self.result_queue = Queue(maxsize=32)
      self.closed = False

  def close(self):
      """关闭工作线程"""
      self.closed = True
      # 清空队列
      while not self.result_queue.empty():
          try:
              self.result_queue.get_nowait()
          except Empty:
              break

  def process_token(self, client, config: Config, tokens: List[int],
                   trim_prefix: int, prompt_wav: str, token_type: str):
      """处理token转音频"""
      try:
          request_id = str(uuid.uuid4())

          # 调用客户端进行token2audio转换
          # 这里需要根据实际的client接口进行调用
          stream = client.token2audio(
              tokens=tokens,
              prompt_wav=prompt_wav,
              token_type=token_type,
              request_id=request_id,
              endpoint=config.endpoint,
              model=config.model,
              direct=config.direct
          )

          # 计算需要裁剪的时长
          trim_duration = 0
          if token_type == "1o" or token_type == "":
              trim_duration = trim_prefix * 24  # 每个token 24ms
          else:
              trim_duration = trim_prefix * 40  # 每个token 40ms

          # 处理流数据
          for frame in stream:
              if self.closed:
                  break
              # 裁剪前缀（这里需要实现音频裁剪逻辑）
              trimmed_frame = self._trim_prefix(frame, trim_duration)
              self.result_queue.put(trimmed_frame)

      except Exception as e:
          print(f"Worker process error: {e}")
      finally:
          # 标记队列结束
          self.result_queue.put(None)

  def _trim_prefix(self, frame: AudioFrame, trim_duration: int) -> AudioFrame:
      """裁剪音频前缀（简化实现）"""
      # 这里需要根据实际音频格式实现裁剪逻辑
      # 简化版本直接返回原frame
      return frame

class Token2Audio:
  """Token转音频主类"""

  def __init__(self, client, config: Config):
      self.client = client
      self.config = config
      self.prefix_token_num = 5
      self.worker_num = 4

  def convert(self, audio_token_channel, voice_id: str,
              callback: Callable[[AudioFrame], None]) -> bool:
      """
      转换token为音频

      Args:
          audio_token_channel: 音频token通道（生成器或迭代器）
          voice_id: 声音ID
          callback: 音频帧回调函数

      Returns:
          bool: 是否成功
      """
      try:
          # 获取提示音频
          prompt_wav = self._get_prompt_wav(voice_id)

          # 初始化变量
          last_tail_tokens = []
          pending_tokens = []
          running_processes = 0
          running_processes_lock = threading.Lock()

          # 工作线程池
          executor = ThreadPoolExecutor(max_workers=self.worker_num)
          workers_queue = Queue(maxsize=self.worker_num)

          # 音频时长统计
          generated_audio_len = 0.0
          first_audio_time = None

          # 结果处理线程
          finished_event = threading.Event()

          def get_token_to_process():
              """获取需要处理的token"""
              nonlocal last_tail_tokens, pending_tokens

              token_to_process = pending_tokens.copy()
              trim_prefix = len(last_tail_tokens)

              if trim_prefix > 0:
                  token_to_process = last_tail_tokens + token_to_process

              if len(token_to_process) > self.prefix_token_num:
                  last_tail_tokens = token_to_process[-self.prefix_token_num:]
              else:
                  last_tail_tokens = []

              pending_tokens = []
              return token_to_process, trim_prefix

          def result_processor():
              """处理工作线程结果"""
              nonlocal generated_audio_len, first_audio_time

              while not finished_event.is_set():
                  try:
                      worker = workers_queue.get(timeout=1.0)
                      if worker is None:
                          break

                      while True:
                          try:
                              frame = worker.result_queue.get(timeout=0.1)
                              if frame is None:  # 工作线程结束标记
                                  break

                              # 计算音频时长
                              duration = len(frame.data) / (frame.channels * 2) / frame.sample_rate
                              generated_audio_len += duration

                              if first_audio_time is None:
                                  first_audio_time = time.time()

                              # 调用回调函数
                              callback(frame)

                          except Empty:
                              continue

                      worker.close()

                  except Empty:
                      continue

          # 启动结果处理线程
          result_thread = threading.Thread(target=result_processor)
          result_thread.start()

          # 定时器，每5秒强制处理一次
          last_process_time = time.time()

          # 主处理循环
          for token_str in audio_token_channel:
              if token_str is None:
                  break

              tokens = self._token_str_to_int(token_str)

              with running_processes_lock:
                  current_running = running_processes

              if current_running > 0:
                  # 有正在处理的请求，加入待处理队列
                  pending_tokens.extend(tokens)
              else:
                  # 没有正在处理的请求，直接处理
                  pending_tokens.extend(tokens)
                  length = len(pending_tokens)

                  should_process = False

                  if first_audio_time is None:
                      # 第一次生成音频，积攒50个token
                      if length >= 50:
                          should_process = True
                  else:
                      # 检查未播放时长
                      unplayed = generated_audio_len - (time.time() - first_audio_time)
                      if unplayed <= 0.5:  # 小于0.5秒就开始生成
                          should_process = True

                  # 定时强制处理
                  if time.time() - last_process_time > 5.0 and len(pending_tokens) > 0:
                      should_process = True

                  if should_process:
                      token_to_process, trim_prefix = get_token_to_process()
                      if token_to_process:
                          worker = Worker()
                          workers_queue.put(worker)

                          with running_processes_lock:
                              running_processes += 1

                          # 提交工作任务
                          future = executor.submit(
                              self._worker_wrapper,
                              worker, token_to_process, trim_prefix, prompt_wav
                          )

                          # 任务完成后减少计数
                          def on_complete(fut):
                              nonlocal running_processes
                              with running_processes_lock:
                                  running_processes -= 1

                          future.add_done_callback(on_complete)
                          last_process_time = time.time()

          # 处理剩余token
          if pending_tokens:
              trim_prefix = len(last_tail_tokens)
              if trim_prefix > 0:
                  pending_tokens = last_tail_tokens + pending_tokens

              worker = Worker()
              workers_queue.put(worker)

              future = executor.submit(
                  self._worker_wrapper,
                  worker, pending_tokens, trim_prefix, prompt_wav
              )
              future.result()  # 等待完成

          # 结束处理
          workers_queue.put(None)  # 结束信号
          finished_event.set()
          result_thread.join()
          executor.shutdown(wait=True)

          return True

      except Exception as e:
          print(f"Convert error: {e}")
          return False

  def _worker_wrapper(self, worker: Worker, tokens: List[int],
                     trim_prefix: int, prompt_wav: str):
      """工作线程包装器"""
      worker.process_token(
          self.client, self.config, tokens,
          trim_prefix, prompt_wav, self.config.token_type
      )

  def _get_prompt_wav(self, voice_id: str) -> str:
      """获取提示音频（需要根据实际实现）"""
      # 这里需要根据voice_id获取对应的提示音频
      return f"prompt_wav_for_{voice_id}"

  def _token_str_to_int(self, token_str: str) -> List[int]:
      """将token字符串转换为整数列表"""
      tokens = []
      curr_num = 0

      for char in token_str:
          if char.isdigit():
              curr_num = curr_num * 10 + int(char)
          else:
              if curr_num != 0:
                  tokens.append(curr_num)
                  curr_num = 0

      # 处理最后一个数字
      if curr_num != 0:
          tokens.append(curr_num)

      return tokens


# 使用示例
def example_usage():
  """使用示例"""

  # 模拟客户端
  class MockClient:
      def token2audio(self, tokens, prompt_wav, token_type, **kwargs):
          # 模拟返回音频流
          for i in range(len(tokens) // 10 + 1):
              yield AudioFrame(
                  data=b"mock_audio_data" * 100,
                  sample_rate=16000,
                  channels=1
              )

  # 配置
  config = Config(
      model="token2audio-vq0206-stream-20241212-zsc",
      token_type="1o"
  )

  # 创建转换器
  client = MockClient()
  converter = Token2Audio(client, config)

  # 模拟token流
  def token_generator():
      test_tokens = [
          "888,265,1189,4502,4264",
          "413,133,3645,4576,5040",
          "692,29,3708,1779,2122"
      ]
      for token_str in test_tokens:
          yield token_str
          time.sleep(0.3)  # 模拟延迟

  # 音频帧回调
  def audio_callback(frame: AudioFrame):
      print(f"Received audio frame: sample_rate={frame.sample_rate}, "
            f"channels={frame.channels}, data_len={len(frame.data)}")

  # 执行转换
  success = converter.convert(
      audio_token_channel=token_generator(),
      voice_id="闫雨婷",
      callback=audio_callback
  )

  print(f"Conversion {'succeeded' if success else 'failed'}")


if __name__ == "__main__":
  example_usage()