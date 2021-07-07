package blog

import (
	"bloga/database"
	"bloga/proto/proto"
	"context"
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection = database.ConnectDB()

type Server struct{}

type BlogItem struct {
	ID      primitive.ObjectID `bson:"id,omitempty"`
	Title   string             `bson:"title"`
	Author  string             `bson:"Author"`
	Content string             `bson:"Content"`
}

func (s *Server) CreateBlog(ctx context.Context, req *proto.CreateBlogReq) (res *proto.CreateBlogRes) {
	blog := req.GetBlog()
	data := BlogItem{
		Author:  blog.GetAuthor(),
		Title:   blog.GetTitle(),
		Content: blog.GetContent(),
	}

	result, err := collection.InsertOne(context.TODO(), data)
	if err != nil {
		log.Fatal(err)
	}
	oid := result.InsertedID.(primitive.ObjectID)

	blog.Id = oid.Hex()

	return &proto.CreateBlogRes{Blog: blog}

}

func (s *Server) ReadBlog(ctx context.Context, req *proto.ReadBlogReq) (*proto.ReadBlogRes, error) {
	// convert string id (from proto) to mongoDB ObjectId
	id, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, errors.New("not found")
	}
	result := collection.FindOne(ctx, bson.M{"id": id})
	// Create an empty BlogItem to write our decode result to
	data := BlogItem{}
	// decode and write to data
	if err := result.Decode(&data); err != nil {
		return nil, errors.New("error while decoding the content")
	}
	// Cast to ReadBlogRes type
	response := &proto.ReadBlogRes{
		Blog: &proto.Blog{
			Id:      id.Hex(),
			Author:  data.Author,
			Title:   data.Title,
			Content: data.Content,
		},
	}
	return response, nil
}

func (s *Server) DeleteBlog(ctx context.Context, req *proto.DeleteBlogReq) (*proto.DeleteBlogRes, error) {
	// Get the ID (string) from the request message and convert it to an Object ID
	id, err := primitive.ObjectIDFromHex(req.GetId())
	// Check for errors
	if err != nil {
		return nil, errors.New("Could not convert to ObjectId:")
	}
	// DeleteOne returns DeleteResult which is a struct containing the amount of deleted docs (in this case only 1 always)
	// So we return a boolean instead
	_, err = collection.DeleteOne(ctx, bson.M{"id": id})
	// Check for errors
	if err != nil {
		return nil, errors.New("Could not find/delete blog ")
	}
	// Return response with success: true if no error is thrown (and thus document is removed)
	return &proto.DeleteBlogRes{
		Success: true,
	}, nil
}

func (s *Server) UpdateBlog(ctx context.Context, req *proto.UpdateBlogReq) (*proto.UpdateBlogRes, error) {
	// Get the blog data from the request
	blog := req.GetBlog()

	// Convert the Id string to a MongoDB ObjectId
	id, err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil {
		return nil, errors.New("Could not convert the supplied blog id to a MongoDB")

	}

	// Convert the data to be updated into an unordered Bson document
	update := bson.M{
		"authord": blog.GetAuthor(),
		"title":   blog.GetTitle(),
		"content": blog.GetContent(),
	}

	// Convert the oid into an unordered bson document to search by id
	filter := bson.M{"id": id}

	// Result is the BSON encoded result
	// To return the updated document instead of original we have to add options.
	result := collection.FindOneAndUpdate(ctx, filter, bson.M{"$set": update}, options.FindOneAndUpdate().SetReturnDocument(1))

	// Decode result and write it to 'decoded'
	decoded := BlogItem{}
	err = result.Decode(&decoded)
	if err != nil {
		return nil, errors.New("Could not find blog with supplied ")

	}
	return &proto.UpdateBlogRes{
		Blog: &proto.Blog{
			Id:      decoded.ID.Hex(),
			Author:  decoded.Author,
			Title:   decoded.Title,
			Content: decoded.Content,
		},
	}, nil
}

func (s *Server) ListBlogs(req *proto.ListBlogReq, stream proto.BlogService_ListBlogsServer) error {
	// Initiate a BlogItem type to write decoded data to
	data := &BlogItem{}
	// collection.Find returns a cursor for our (empty) query
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return errors.New("can not find")
	}
	// An expression with defer will be called at the end of the function
	defer cursor.Close(context.Background())
	// cursor.Next() returns a boolean, if false there are no more items and loop will break
	for cursor.Next(context.Background()) {
		// Decode the data at the current pointer and write it to data
		err := cursor.Decode(data)
		// check error
		if err != nil {
			return errors.New("Could not decode data")
		}
		// If no error is found send blog over stream
		stream.Send(&proto.ListBlogRes{
			Blog: &proto.Blog{
				Id:      data.ID.Hex(),
				Author:  data.Author,
				Content: data.Content,
				Title:   data.Title,
			},
		})
	}
	// Check if the cursor has any errors
	if err := cursor.Err(); err != nil {
		return errors.New("Unkown cursor error")
	}
	return nil
}
