package models

import "time"
import "encoding/json"

// Layout represents a user-defined viewing grid or arrangement
type Layout struct {
	ID      int64             `json:"id"`
	UserID  int64             `json:"user_id"`
	Name    string            `json:"name"`
	Mode    string            `json:"mode"`
	Payload json.RawMessage   `json:"payload"` // Raw JSON bytes from SQLite
	Items   []LayoutItem      `json:"items"`   // Populated via SQL JOIN or secondary query
	CreatedAt time.Time `json:"created_at"`
}

// LayoutItem represents an individual block within a layout
type LayoutItem struct {
	ID       int64           `json:"-"` // Hidden from frontend
	LayoutID int64           `json:"-"` // Hidden from frontend
	Type     string          `json:"type"`
	Payload  json.RawMessage `json:"payload"` // Raw JSON bytes from SQLite
}

//  layout api
// {
//   "data": {
//     "layouts": [
//       {
//         "id": 1,
//         "name": "3F",
//         "mode": "grid",
//         "payload": {
//           "rows": 3,
//           "cols": 3
//         },
//         "items": [
//           {
//             "type": "camera",
//             "payload": {
//               "camera_id": 1
//             }
//           },
//           {
//             "type": "roi",
//             "payload": {
//               "camera_id": 1,
//               "x": 0.2,
//               "y": 0.3,
//               "width": 0.3,
//               "height": 0.3
//             }
//           }
//         ]
//       }
//     ]
//   },
//   "message": "success"
// }