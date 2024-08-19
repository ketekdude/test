package main

import (
	"context"
	"io"
	"log"
	"os"
	"time"

	pb "test/grpc-stream/proto"

	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"

	// Adjust the size for which a large file will be broken
	// down into multiple parts during a stream request
	chunkSize = 5 * MB
)

const MB = 1 << 20

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewUploaderClient(conn)

	if len(os.Args) == 1 {
		log.Fatal("need filename argument!")
	}
	// fileName := os.Args[1]

	// uploadTook := testAvg(func() time.Duration { return Upload(c, fileName) })
	streamTook := testAvg(func() time.Duration { return DownloadStream(c) })
	uploadTook := testAvg(func() time.Duration { return Download(c) })
	log.Printf(" Unary Avg: %s", uploadTook)
	log.Printf("Stream Avg: %s", streamTook)
}

func testAvg(fn func() time.Duration) time.Duration {
	const N = 10
	var total time.Duration
	for i := 0; i < N; i++ {
		total += fn()
	}
	return total / N
}

func Upload(client pb.UploaderClient, fileName string) time.Duration {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fh, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer fh.Close()

	stat, err := fh.Stat()
	if err != nil {
		log.Fatal(err)
	}
	size := stat.Size()

	log.Printf("[Upload] Will send file %q with size %d\n", fh.Name(), size)

	start := time.Now()

	data, err := io.ReadAll(fh)
	if err != nil {
		log.Fatal(err)
	}
	sizePB := &pb.FileHeader_FileSize{
		FileSize: size,
	}
	_, err = client.Upload(ctx, &pb.FileTransferRequest{
		Header: &pb.FileHeader{
			Name:             fh.Name(),
			OptionalFileSize: sizePB,
		},
		Data: data,
	})
	if err != nil {
		log.Fatal(err)
	}

	took := time.Since(start)
	log.Printf("[Upload] Took: %s\n", took)
	return took
}

func UploadStream(client pb.UploaderClient, fileName string) time.Duration {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fh, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer fh.Close()

	stat, err := fh.Stat()
	if err != nil {
		log.Fatal(err)
	}
	size := stat.Size()
	log.Printf("[UploadStream] Will stream file %q with size %d\n", fh.Name(), size)

	start := time.Now()

	stream, err := client.UploadStream(ctx)
	if err != nil {
		log.Fatal(err)
	}
	sizePB := &pb.FileHeader_FileSize{
		FileSize: size,
	}
	header := &pb.FileHeader{Name: fh.Name(), OptionalFileSize: sizePB}
	err = stream.Send(&pb.FileStreamRequest{Contents: &pb.FileStreamRequest_Header{Header: header}})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("  Sent header. Now sending data chunks...")

	buf := make([]byte, chunkSize)
	chunkCount := 0
	for {
		n, err := fh.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		//fmt.Printf("Sending chunk #%d with size %d\n", i, n)
		err = stream.Send(&pb.FileStreamRequest{Contents: &pb.FileStreamRequest_Chunk{Chunk: buf[:n]}})
		chunkCount++
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
	}
	stream.CloseAndRecv()

	took := time.Since(start)
	log.Printf("  Sent %d chunk(s)\n", chunkCount)
	log.Printf("[UploadStream] Took: %s\n", took)
	return took
}

func DownloadStream(client pb.UploaderClient) time.Duration {
	start := time.Now()
	// Create a context with timeout (if needed)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Call the streaming RPC
	stream, err := client.DownloadStream(ctx, &pb.FileRequest{FileUrl: "http://example.com/file.pdf"})
	if err != nil {
		log.Fatalf("failed to start streaming: %v", err)
	}
	chunkCount := 0
	// Process the streamed chunks
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("stream error: %v", err)
		}

		chunkCount++
		// Process each chunk
		log.Printf("Received chunk %d with size %d bytes", chunk.ChunkNumber, len(chunk.ChunkData))
	}
	took := time.Since(start)
	log.Printf("  Sent %d chunk(s)\n", chunkCount)
	log.Printf("[DownloadStream] Took: %s\n", took)
	return took
}

func Download(client pb.UploaderClient) time.Duration {
	start := time.Now()
	// Create a context with timeout (if needed)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Call the streaming RPC
	_, err := client.Download(ctx, &pb.FileRequest{FileUrl: "http://example.com/file.pdf"})
	if err != nil {
		log.Fatalf("failed to start streaming: %v", err)
	}

	took := time.Since(start)
	log.Printf("[Download] Took: %s\n", took)
	return took
}
