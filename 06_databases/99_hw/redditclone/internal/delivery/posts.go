package delivery

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gitlab.vk-golang.com/vk-golang/lectures/06_databases/99_hw/redditclone/internal/di"
	"gitlab.vk-golang.com/vk-golang/lectures/06_databases/99_hw/redditclone/internal/domain"
	"gitlab.vk-golang.com/vk-golang/lectures/06_databases/99_hw/redditclone/internal/lib"
	"gitlab.vk-golang.com/vk-golang/lectures/06_databases/99_hw/redditclone/internal/service"
	"go.uber.org/zap"
)

type PostsHandler struct {
	logger       *zap.SugaredLogger
	postsService di.PostsService
}

func NewPostsHandler(svc di.PostsService, logger *zap.SugaredLogger) *PostsHandler {
	return &PostsHandler{
		postsService: svc,
		logger:       logger,
	}
}

func (h *PostsHandler) SetupRoutes(pub, pr *mux.Router) (*mux.Router, *mux.Router) {
	pub.HandleFunc("/posts/", h.GetPosts).Methods(http.MethodGet)
	pub.HandleFunc("/posts/{category}", h.GetPostByCategory).Methods(http.MethodGet)
	pub.HandleFunc("/post/{id}", h.GetPostByID).Methods(http.MethodGet)
	pub.HandleFunc("/user/{username}", h.PostsByUsername).Methods(http.MethodGet)

	pr.HandleFunc("/post/{postID}/{commentID}", h.DeleteComment).Methods(http.MethodDelete)
	pr.HandleFunc("/post/{postID}/upvote", h.UpvotePost).Methods(http.MethodGet)
	pr.HandleFunc("/post/{postID}/downvote", h.DownvotePost).Methods(http.MethodGet)
	pr.HandleFunc("/post/{postID}/unvote", h.UnvotePost).Methods(http.MethodGet)
	pr.HandleFunc("/posts", h.CreatePost).Methods(http.MethodPost)
	pr.HandleFunc("/post/{id}", h.AddComment).Methods(http.MethodPost)
	return pub, pr
}

func (h *PostsHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := h.postsService.GetAllPosts(r.Context())
	if err != nil {
		lib.WriteError(w, err)
		return
	}

	var postsResponse []PostResponse
	for _, post := range posts {
		postsResponse = append(postsResponse, toPostResponse(post))
	}

	success, _ := json.Marshal(&postsResponse)
	w.Write(success)
}

func (h *PostsHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var req PostDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, err)
		return
	}

	post, err := h.postsService.CreatePost(r.Context(), service.CreatePostDTO{
		Title:    req.Title,
		Text:     req.Text,
		Category: req.Category,
		Type:     domain.PostType(req.Type),
		URL:      req.URL,
	})

	if err != nil {
		lib.WriteError(w, err)
		return
	}

	postResponse := toPostResponse(post)
	success, _ := json.Marshal(&postResponse)
	w.Write(success)
}

func (h *PostsHandler) GetPostByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		lib.WriteError(w, err)
		return
	}

	post, err := h.postsService.GetPostByID(r.Context(), uint64(id))

	if err != nil {
		lib.WriteError(w, err)
		return
	}

	postResponse := toPostResponse(post)
	success, _ := json.Marshal(&postResponse)
	w.Write(success)
}

func (h *PostsHandler) GetPostByCategory(w http.ResponseWriter, r *http.Request) {
	category := mux.Vars(r)["category"]

	posts, err := h.postsService.GetPostByCategory(r.Context(), category)
	if err != nil {
		lib.WriteError(w, err)
		return
	}

	var postsResponse []PostResponse
	for _, post := range posts {
		postsResponse = append(postsResponse, toPostResponse(post))
	}

	success, _ := json.Marshal(&postsResponse)
	w.Write(success)
}

type AddCommentReq struct {
	Body string `json:"comment"`
}

func (h *PostsHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	h.logger.Infoln(r.URL)
	var req AddCommentReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		lib.WriteError(w, err)
		return
	}
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		h.logger.Infoln(mux.Vars(r))

		lib.WriteError(w, err)
		return
	}

	post, err := h.postsService.AddCommentToPost(r.Context(), uint64(id), req.Body)
	if err != nil {
		lib.WriteError(w, err)
		return
	}

	postResponse := toPostResponse(post)
	success, _ := json.Marshal(&postResponse)
	w.Write(success)
}

func (h *PostsHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.Atoi(mux.Vars(r)["postID"])
	if err != nil {
		lib.WriteError(w, err)
		return
	}
	commentID, err := strconv.Atoi(mux.Vars(r)["commentID"])
	if err != nil {
		lib.WriteError(w, err)
		return
	}

	post, err := h.postsService.DeleteComment(r.Context(), uint64(postID), uint64(commentID))
	if err != nil {
		lib.WriteError(w, err)
		return
	}

	postResponse := toPostResponse(post)
	success, _ := json.Marshal(&postResponse)
	w.Write(success)
}

func (h *PostsHandler) UpvotePost(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.Atoi(mux.Vars(r)["postID"])
	if err != nil {
		lib.WriteError(w, err)
		return
	}

	post, err := h.postsService.Upvote(r.Context(), uint64(postID))
	if err != nil {
		lib.WriteError(w, err)
		return
	}

	postResponse := toPostResponse(post)
	success, _ := json.Marshal(&postResponse)
	w.Write(success)
}
func (h *PostsHandler) DownvotePost(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.Atoi(mux.Vars(r)["postID"])
	if err != nil {
		lib.WriteError(w, err)
		return
	}

	post, err := h.postsService.Downvote(r.Context(), uint64(postID))
	if err != nil {
		lib.WriteError(w, err)
		return
	}

	postResponse := toPostResponse(post)
	success, _ := json.Marshal(&postResponse)
	w.Write(success)
}
func (h *PostsHandler) UnvotePost(w http.ResponseWriter, r *http.Request) {
	postID, err := strconv.Atoi(mux.Vars(r)["postID"])
	if err != nil {
		lib.WriteError(w, err)
		return
	}

	post, err := h.postsService.DiscardVote(r.Context(), uint64(postID))
	if err != nil {
		lib.WriteError(w, err)
		return
	}

	postResponse := toPostResponse(post)
	success, _ := json.Marshal(&postResponse)
	w.Write(success)
}

func (h *PostsHandler) PostsByUsername(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	posts, err := h.postsService.GetPostsByUsername(r.Context(), username)
	if err != nil {
		lib.WriteError(w, err)
		return
	}

	var postsResponse []PostResponse
	for _, post := range posts {
		postsResponse = append(postsResponse, toPostResponse(post))
	}

	success, _ := json.Marshal(&postsResponse)
	w.Write(success)
}
