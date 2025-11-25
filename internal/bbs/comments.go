package bbs

// AddComment adds a comment to a post.
func (b *BBS) AddComment(boardName string, postID int, author, content string, parentID int) (*Comment, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	board, ok := b.boards[boardName]
	if !ok {
		return nil, ErrBoardNotFound
	}

	board.mu.Lock()
	defer board.mu.Unlock()

	// Find the post
	var post *Post
	for i := range board.posts {
		if board.posts[i].ID == postID {
			post = &board.posts[i]
			break
		}
	}
	if post == nil {
		return nil, ErrPostNotFound
	}

	// Create comment
	commentID := len(post.Comments) + 1
	comment := Comment{
		ID:        commentID,
		PostID:    postID,
		ParentID:  parentID,
		Author:    author,
		Content:   content,
		CreatedAt: b.now(),
	}

	post.Comments = append(post.Comments, comment)

	// Save to disk
	if b.posts != nil {
		if err := b.posts.Save(boardName, board.posts); err != nil {
			return nil, err
		}
	}

	return &comment, nil
}

// ListComments returns all comments for a post.
func (b *BBS) ListComments(boardName string, postID int) ([]Comment, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	board, ok := b.boards[boardName]
	if !ok {
		return nil, ErrBoardNotFound
	}

	board.mu.RLock()
	defer board.mu.RUnlock()

	for _, post := range board.posts {
		if post.ID == postID {
			return post.Comments, nil
		}
	}

	return nil, ErrPostNotFound
}
