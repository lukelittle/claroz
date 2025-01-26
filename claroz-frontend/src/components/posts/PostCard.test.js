import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { BrowserRouter } from 'react-router-dom';
import PostCard from './PostCard';
import { postAPI } from '../../services/api';

// Mock the API
jest.mock('../../services/api', () => ({
  postAPI: {
    likePost: jest.fn(),
    unlikePost: jest.fn(),
  },
}));

// Mock the useAuth hook
jest.mock('../../context/AuthContext', () => ({
  useAuth: () => ({
    user: { id: 'test-user-id' },
    isAuthenticated: true,
  }),
}));

const mockPost = {
  id: 'test-post-id',
  content: 'Test post content',
  imageUrl: 'test-image.jpg',
  user: {
    id: 'test-user-id',
    username: 'testuser',
    avatar: 'test-avatar.jpg',
  },
  likes: [],
  comments: [],
  createdAt: new Date().toISOString(),
};

const renderPostCard = () => {
  return render(
    <BrowserRouter>
      <PostCard post={mockPost} />
    </BrowserRouter>
  );
};

describe('PostCard', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  test('renders post content', () => {
    renderPostCard();
    
    expect(screen.getByText(mockPost.content)).toBeInTheDocument();
    expect(screen.getByText(mockPost.user.username)).toBeInTheDocument();
    expect(screen.getByAltText('post')).toHaveAttribute('src', mockPost.imageUrl);
    expect(screen.getByAltText('avatar')).toHaveAttribute('src', mockPost.user.avatar);
  });

  test('handles like/unlike', async () => {
    postAPI.likePost.mockResolvedValueOnce({});
    postAPI.unlikePost.mockResolvedValueOnce({});
    
    renderPostCard();
    
    const likeButton = screen.getByRole('button', { name: /like/i });
    
    // Like
    await userEvent.click(likeButton);
    expect(postAPI.likePost).toHaveBeenCalledWith(mockPost.id);
    
    // Unlike
    await userEvent.click(likeButton);
    expect(postAPI.unlikePost).toHaveBeenCalledWith(mockPost.id);
  });

  test('navigates to user profile on username click', () => {
    renderPostCard();
    
    const usernameLink = screen.getByText(mockPost.user.username);
    expect(usernameLink.closest('a')).toHaveAttribute('href', `/profile/${mockPost.user.id}`);
  });

  test('displays timestamp', () => {
    renderPostCard();
    
    // Check if timestamp is displayed in some format
    expect(screen.getByText(/ago$/i)).toBeInTheDocument();
  });

  test('shows like count', () => {
    const postWithLikes = {
      ...mockPost,
      likes: [{ id: 'like1' }, { id: 'like2' }],
    };
    
    render(
      <BrowserRouter>
        <PostCard post={postWithLikes} />
      </BrowserRouter>
    );
    
    expect(screen.getByText('2')).toBeInTheDocument();
  });
});
