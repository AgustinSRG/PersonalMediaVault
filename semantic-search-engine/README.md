# Semantic Search Engine

The semantic search is responsible for the following features for PersonalMediaVault:

 - Semantic search of images / titles: Instead of searching with tags or exact words in the title, semantic search allows the user to perform search in natural language related to the meaning of the title or the image. Example: If the user searches `Black Cat`, images of blacks cats in their vault will appear first.
 - Search by image similarity: The user selects an image, and similar images will appear first.

In order to achieve this functionality, a multi-modal embedding model is needed, which supports both images and text. This embedding model will turn both images and text into vectors of the same size. Then, by sorting vectors by distance, semantic search is achieved.

The model type chosen for this task is [OpenCLIP](https://github.com/mlfoundations/open_clip), an open source version of the original CLIP model, developed by OpenAI.

The model inference is achieved using the [ONNX Runtime](https://github.com/microsoft/onnxruntime), so models need to be adapted for this runtime.

## Internal server

The inference runs in an internal Rust server that exposes a GRPC service for both text and image embeddings.

The source code for this server can be found in the [server](./server/) folder.

## Protocol

The communication protocol between the PersonalMediaVault backend and the server is GRPC, using protocol buffers.

You can find the protocol files in the [protocol](./protocol/) folder.
