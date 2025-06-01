#!/usr/bin/env python3

import yaml
import argparse
import sys
import re
from collections import deque


def load_yaml(file_path):
    """加载YAML文件"""
    try:
        with open(file_path, 'r', encoding='utf-8') as file:
            return yaml.safe_load(file)
    except Exception as e:
        print(f"Error loading YAML file: {e}")
        sys.exit(1)


def save_yaml(data, output_path):
    """保存YAML文件"""
    try:
        with open(output_path, 'w', encoding='utf-8') as file:
            yaml.dump(data, file, sort_keys=False, default_flow_style=False)
        print(f"Cleaned OpenAPI spec saved to {output_path}")
    except Exception as e:
        print(f"Error saving YAML file: {e}")
        sys.exit(1)


def extract_refs_from_object(obj):
    """从对象中提取所有$ref引用"""
    refs = set()

    def extract_refs(item):
        if isinstance(item, dict):
            for key, value in item.items():
                if key == '$ref' and isinstance(value, str):
                    refs.add(value)
                elif isinstance(value, (dict, list)):
                    extract_refs(value)
        elif isinstance(item, list):
            for element in item:
                extract_refs(element)

    extract_refs(obj)
    return refs


def get_component_path_from_ref(ref):
    """从$ref引用中提取组件路径"""
    if not ref.startswith('#/components/'):
        return None

    # 移除开头的 #/
    path = ref[2:].split('/')
    return path


def clean_openapi(spec, keep_paths):
    """清理OpenAPI规范，只保留指定路径及其依赖"""
    if 'paths' not in spec:
        print("No paths found in the OpenAPI spec")
        return spec

    # 创建一个新的规范，保留基本信息
    cleaned_spec = {
        key: value for key, value in spec.items()
        if key not in ['paths', 'components']
    }

    # 只保留指定的路径
    cleaned_paths = {}
    for path in keep_paths:
        if path in spec['paths']:
            cleaned_paths[path] = spec['paths'][path]
        else:
            print(f"Warning: Path '{path}' not found in the OpenAPI spec")

    cleaned_spec['paths'] = cleaned_paths

    # 找出所有引用的组件
    all_refs = set()
    for path_item in cleaned_paths.values():
        refs = extract_refs_from_object(path_item)
        all_refs.update(refs)

    # 递归查找所有依赖的组件
    processed_refs = set()
    refs_to_process = deque(all_refs)

    while refs_to_process:
        current_ref = refs_to_process.popleft()
        if current_ref in processed_refs:
            continue

        processed_refs.add(current_ref)
        component_path = get_component_path_from_ref(current_ref)

        if component_path:
            # 获取引用的组件
            component = spec
            for part in component_path:
                if part in component:
                    component = component[part]
                else:
                    component = None
                    break

            if component:
                # 查找该组件中的引用
                component_refs = extract_refs_from_object(component)
                for ref in component_refs:
                    if ref not in processed_refs:
                        refs_to_process.append(ref)

    # 构建清理后的组件
    if 'components' in spec:
        cleaned_components = {}

        for ref in processed_refs:
            component_path = get_component_path_from_ref(ref)
            if not component_path:
                continue

            # 确保组件类型存在
            component_type = component_path[1]
            component_name = component_path[2]

            if component_type not in cleaned_components:
                cleaned_components[component_type] = {}

            # 从原始规范中获取组件
            if (component_type in spec['components'] and
                component_name in spec['components'][component_type]):
                cleaned_components[component_type][component_name] = spec['components'][component_type][component_name]

        if cleaned_components:
            cleaned_spec['components'] = cleaned_components

    return cleaned_spec


def main():
    parser = argparse.ArgumentParser(description='Clean OpenAPI specification by keeping only specified paths and their dependencies.')
    parser.add_argument('input_file', help='Path to the input OpenAPI YAML file')
    parser.add_argument('output_file', help='Path to save the cleaned OpenAPI YAML file')
    parser.add_argument('--paths', nargs='+', required=True, help='List of paths to keep (e.g. /assistants /assistants/{assistant_id})')

    args = parser.parse_args()

    # 加载OpenAPI规范
    spec = load_yaml(args.input_file)

    # 清理规范
    cleaned_spec = clean_openapi(spec, args.paths)

    # 保存清理后的规范
    save_yaml(cleaned_spec, args.output_file)


if __name__ == "__main__":
    main()


# python clean_openapi.py \
#     openapi.yaml \
#     cleaned_openapi.yaml \
#     --paths "/responses" "/responses/{response_id}" "/responses/{response_id}/input_items"
