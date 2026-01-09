package firebase

import (
	"context"
	"log/slog"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
)

// InitFirestore initializes a Firebase App using Application Default Credentials
// and returns a Firestore client.
//
// The function expects credentials to be available via one of the following:
//   - GOOGLE_APPLICATION_CREDENTIALS environment variable
//   - Default credentials in a Google Cloud environment (Cloud Run, GKE, etc.)
//
// The caller is responsible for calling client.Close() when shutting down
// the application.
func InitFirestore(ctx context.Context) (*firestore.Client, error) {
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		slog.Error(
			"failed to initialize Firebase app",
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	slog.Info("Firebase app initialized")

	client, err := app.Firestore(ctx)
	if err != nil {
		slog.Error(
			"failed to initialize Firestore client",
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	slog.Info("Firestore client connected")

	return client, nil
}
