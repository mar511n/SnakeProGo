package game

import (
	"fmt"
	"os/exec"

	"github.com/hajimehoshi/ebiten/v2"
)

// RenderVideo creates a video from a sequence of images and frame indices.
// images: A slice of ebiten.Image to be used as frames.
// frameIndices: A slice of frame indices corresponding to each image.
// fps: The frames per second for the output video.
// filename: The name of the output video file.
func RenderVideo(numImages, width, height int, images chan *ebiten.Image, frameIndices chan int, fps int, filename string) error {
	// Check if ffmpeg is installed
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not found: %w", err)
	}

	// Prepare ffmpeg command
	cmd := exec.Command("ffmpeg",
		"-y",             // Overwrite output file without asking
		"-f", "rawvideo", // Input format as raw video
		"-pix_fmt", "rgba", // Input pixel format (ebiten uses RGBA)
		"-s", fmt.Sprintf("%dx%d", width, height), // Video size
		"-r", fmt.Sprintf("%d", fps), // Input framerate
		"-i", "-", // Input from stdin
		"-c:v", "libx264", // Video codec
		"-pix_fmt", "yuv420p", // Output pixel format for compatibility
		filename, // Output filename
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	// Start ffmpeg process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Buffer for pixel data matches image size * 4 (RGBA)
	// We allocate inside the Goroutine to avoid race condition if called concurrently,
	// but here we are single threaded.
	pixels := make([]byte, width*height*4)

	// We use a separate goroutine or just write sequentially.
	// Since ffmpeg reads from stdin, we can write sequentially.
	// IMPORTANT: Ensure we close stdin when done, or if an error occurs.

	// Use a cleanup function to ensure stdin is closed and process waited for
	cleanup := func() error {
		stdin.Close()
		if err := cmd.Wait(); err != nil {
			return fmt.Errorf("ffmpeg finished with error: %w", err)
		}
		return nil
	}

	lastIndex := -1
	currIndex := 0
	var img *ebiten.Image
	for i := 0; i < numImages; i++ {
		img = <-images
		currIndex = <-frameIndices
		if img == nil {
			continue
		}

		// Calculate number of frames to write for this image
		duration := 0
		if i < numImages-1 {
			duration = currIndex - lastIndex
		} else {
			// Last image shown for 10 frames
			duration = 10
		}

		if duration <= 0 {
			continue
		}

		// Read pixels from image
		// ReadPixels fills the provided byte slice with RGBA values.
		img.ReadPixels(pixels)

		// Write the frame 'duration' times
		for j := 0; j < duration; j++ {
			if _, err := stdin.Write(pixels); err != nil {
				stdin.Close() // Close stdin to convert error
				cmd.Wait()    // Wait to cleanup
				return fmt.Errorf("failed to write frame to ffmpeg: %w", err)
			}
		}

		lastIndex = currIndex
	}

	return cleanup()
}
