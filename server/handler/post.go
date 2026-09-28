package handler

import (
	"context"

	pb "github.com/kidrury/grpc-blog/gen/blog"
	"github.com/kidrury/grpc-blog/internal/domain"
	"github.com/kidrury/grpc-blog/internal/service"
)

// type UserService interface {
// 	CreateUser()
// }

func domainToProtoPost(post *domain.Post) *pb.Post {
	return &pb.Post{
		Id:        post.ID,
		Title:     post.Title,
		Content:   post.Content,
		CreatedBy: post.CreatedBy,
		CreatedAt: post.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

type PostHandler struct {
	pb.UnimplementedBlogServiceServer
	service *service.PostService
}

func NewPostHandler(service *service.PostService) *PostHandler {
	return &PostHandler{
		service: service,
	}
}

func (h *PostHandler) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.CreatePostResponse, error) {
	//get claims

	//slog here

	post, err := h.service.CreatePost(ctx, req.Title, req.Content)
	if err != nil {
		return nil, mapErr(err)
	}

	return &pb.CreatePostResponse{
		Post: domainToProtoPost(post),
	}, nil
}

func (h *PostHandler) GetPost(ctx context.Context, req *pb.GetPostRequest) (*pb.GetPostResponse, error) {
	//get claims

	//slog here

	post, err := h.service.GetPostByID(ctx, req.Id)
	if err != nil {
		return nil, mapErr(err)
	}

	return &pb.GetPostResponse{
		Post: domainToProtoPost(post),
	}, nil
}

func (h *PostHandler) ListPosts(ctx context.Context, req *pb.ListPostsRequest) (*pb.ListPostsResponse, error) {
	//get claims

	//slog here

	posts, err := h.service.ListPosts(ctx)
	if err != nil {
		return nil, mapErr(err)
	}

	pbPosts := make([]*pb.Post, 0, len(posts))

	for _, post := range posts {
		pbPosts = append(pbPosts, domainToProtoPost(post))
	}

	return &pb.ListPostsResponse{
		Posts: pbPosts,
	}, nil
}

func (h *PostHandler) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*pb.DeletePostResponse, error) {
	err := h.service.DeletePost(ctx, req.Id)

	if err != nil {
		return nil, mapErr(err)
	}

	return &pb.DeletePostResponse{
		Deleted: true,
		Message: "post deleted",
	}, nil
}
