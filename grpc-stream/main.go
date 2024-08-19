//go:generate ./generate_proto.sh
package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	pb "test/grpc-stream/proto"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

const (
	port       = ":50051"
	maxReceive = 500 * MB
)

const MB = 1 << 20

type server struct {
	pb.UploaderServer
}

func (s *server) Upload(ctx context.Context, req *pb.FileTransferRequest) (*empty.Empty, error) {
	var (
		header *pb.FileHeader
		buf    bytes.Buffer
	)

	start := time.Now()

	header = req.GetHeader()
	log.Printf("[Upload] Got file request header: %s\n", header.GetName())
	if header.OptionalFileSize != nil {
		log.Printf("  Reported file size should be: %d\n", header.GetFileSize())
	}

	if data := req.GetData(); data != nil {
		buf.Write(data)
	}

	took := time.Since(start)
	log.Printf("  Total bytes received: %d\n", buf.Len())
	log.Printf("[Upload] Took: %s\n", took)
	return &empty.Empty{}, nil
}

func (s *server) UploadStream(stream pb.Uploader_UploadStreamServer) error {
	var (
		header     *pb.FileHeader
		buf        bytes.Buffer
		chunkCount int
	)
	start := time.Now()
	for {
		req, err := stream.Recv()
		if err == io.EOF || stream.Context().Err() != nil {
			break
		}
		if err != nil {
			log.Println(err)
			stream.SendAndClose(&empty.Empty{})
			return err
		}

		if req.GetHeader() != nil {
			header = req.GetHeader()
			log.Printf("[UploadStream] Got file request header: %s\n", header.GetName())
			if header.OptionalFileSize != nil {
				log.Printf("  Reported file size should be: %d\n", header.GetFileSize())
			}
			continue
		}

		if req.GetChunk() != nil {
			chunkCount++
			buf.Write(req.GetChunk())
			//fmt.Printf("Got file data chunk #%d of size %d\n", chunkCount, n)
		}
	}
	stream.SendAndClose(&empty.Empty{})
	took := time.Since(start)
	log.Printf("  Total bytes received (in %d chunk(s)): %d\n", chunkCount, buf.Len())
	log.Printf("[UploadStream] Took: %s\n", took)
	return nil
}

func (s *server) DownloadStream(req *pb.FileRequest, stream pb.Uploader_DownloadStreamServer) error {
	// fileURL := req.GetFileUrl()
	var (
		buf bytes.Buffer
	)
	fileURL := "https://images.pexels.com/photos/27351031/pexels-photo-27351031/free-photo-of-essaouira-view.jpeg"
	// Perform an HTTP GET request to the file URL
	resp, err := http.Get(fileURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	start := time.Now()
	// Define chunk size (e.g., 4KB)
	const chunkSize = 1024 * 100
	buffer := make([]byte, chunkSize)
	chunkNumber := 0

	// Stream the file in chunks
	for {
		// Read a chunk from the file
		n, err := resp.Body.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Create a FileChunk message
		chunk := &pb.FileChunk{
			ChunkData:   buffer[:n], // Send only the bytes read
			ChunkNumber: int32(chunkNumber),
		}

		// Send the chunk to the client
		if err := stream.Send(chunk); err != nil {
			return err
		}
		buf.Write(buffer)
		chunkNumber++
	}
	took := time.Since(start)
	log.Printf("  Total bytes sent (in %d chunk(s)): %d\n", chunkNumber, buf.Len())
	log.Printf("[DownloadStream] Took: %s\n", took)
	return nil
}

func (s *server) Download(ctx context.Context, req *pb.FileRequest) (*pb.FileData, error) {
	fileURL := "https://images.pexels.com/photos/27351031/pexels-photo-27351031/free-photo-of-essaouira-view.jpeg"
	// Perform an HTTP GET request to the file URL
	resp, err := http.Get(fileURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	start := time.Now()

	// Read the entire file into memory
	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Create and return a FileResponse with the entire file data

	took := time.Since(start)
	log.Printf("[Download] Took: %s\n", took)
	return &pb.FileData{
		ChunkData: fileData,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(grpc.MaxRecvMsgSize(maxReceive))
	pb.RegisterUploaderServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// func (s *server) DownloadStream(req *pb.FileRequest, stream pb.Uploader_DownloadStreamServer) error {
// 	// fileURL := req.GetFileUrl()
// 	var (
// 		buf bytes.Buffer
// 	)
// 	fileURL := "https://images.pexels.com/photos/27351031/pexels-photo-27351031/free-photo-of-essaouira-view.jpeg"
// 	// Perform an HTTP GET request to the file URL
// 	start := time.Now()
// 	// Define chunk size (100KB)
// 	const chunkSize = 100 * 1024 // 100 KB in bytes
// 	const numWorkers = 10        // Number of concurrent workers
// 	chunkNumber := 0
// 	var wg sync.WaitGroup
// 	ch := make(chan *pb.FileChunk)

// 	// Calculate total file size (assuming you can get the size)
// 	resp, err := http.Head(fileURL)
// 	if err != nil {
// 		return err
// 	}
// 	fileSize := resp.Header.Get("Content-Length")
// 	if fileSize == "" {
// 		return grpc.Errorf(codes.InvalidArgument, "Unable to get file size")
// 	}

// 	totalSize, err := strconv.Atoi(fileSize)
// 	if err != nil {
// 		return grpc.Errorf(codes.InvalidArgument, "Invalid file size")
// 	}

// 	// Create workers
// 	for i := 0; i < numWorkers; i++ {
// 		start := i * chunkSize
// 		if start >= totalSize {
// 			break
// 		}
// 		chunkNumber++
// 		wg.Add(1)
// 		go workerDownloadStream(fileURL, int64(start), chunkSize, &wg, ch)
// 	}

// 	// Create a goroutine to close the channel once all workers are done
// 	go func() {
// 		wg.Wait()
// 		close(ch)
// 	}()

// 	// Send chunks to the client
// 	for chunk := range ch {
// 		if err := stream.Send(chunk); err != nil {
// 			return err
// 		}
// 	}
// 	took := time.Since(start)
// 	log.Printf("  Total bytes sent (in %d chunk(s)): %d\n", chunkNumber, buf.Len())
// 	log.Printf("[DownloadStream] Took: %s\n", took)
// 	return nil
// }

// // Worker function to read chunks and send them to the channel
// func workerDownloadStream(fileURL string, start int64, chunkSize int, wg *sync.WaitGroup, ch chan<- *pb.FileChunk) {
// 	defer wg.Done()

// 	// Perform an HTTP GET request to the file URL
// 	resp, err := http.Get(fileURL)
// 	if err != nil {
// 		log.Printf("Error fetching file: %v", err)
// 		return
// 	}
// 	defer resp.Body.Close()

// 	// Create a buffer to hold the chunk data
// 	buffer := make([]byte, chunkSize)

// 	// Read the file from the start position
// 	if _, err := io.CopyN(io.Discard, resp.Body, start); err != nil && err != io.EOF {
// 		log.Printf("Error skipping file: %v", err)
// 		return
// 	}

// 	// Read and send the chunk data
// 	n, err := io.ReadFull(resp.Body, buffer)
// 	if err == io.EOF && n == 0 {
// 		return
// 	}
// 	if err != nil && err != io.EOF {
// 		log.Printf("Error reading file: %v", err)
// 		return
// 	}

// 	chunk := &pb.FileChunk{
// 		ChunkData:   buffer[:n], // Send only the bytes read
// 		ChunkNumber: int32(start / int64(chunkSize)),
// 	}

// 	// Send the chunk to the channel
// 	ch <- chunk
// }
