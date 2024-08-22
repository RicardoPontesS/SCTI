package database

import (
  "fmt"
  "database/sql"
  "errors"

  _ "github.com/lib/pq"
)

type Activity struct {
  Activity_id int
  Spots int
  Activity_type string
  Room string
  Speaker string
  Topic string
  Description string
  Time string
  Day int
}

func GetAllActivities() (activities []Activity, err error) {
  query := `
  SELECT *
  FROM activities
  `

  rows, err := DB.Query(query)
  if err != nil {
    return nil, fmt.Errorf("could not retrieve activities: %v", err)
  }
  defer rows.Close()

  for rows.Next() {
    var a Activity
    err := rows.Scan(
      &a.Activity_id,
      &a.Spots,
      &a.Activity_type,
      &a.Room,
      &a.Speaker,
      &a.Topic,
      &a.Description,
      &a.Time,
      &a.Day,
      )

    if err != nil {
      return nil, fmt.Errorf("could not scan activity: %v", err)
    }
    activities = append(activities, a)
  }

  if err = rows.Err(); err != nil {
    return nil, fmt.Errorf("row iteration error: %v", err)
  }

  return activities, nil
}

func GetActivity(id int) (a Activity, err error) {
  query := `
  SELECT *
  FROM activities
  WHERE activities.id = $1
  `

  err = DB.QueryRow(query, id).Scan(
    &a.Activity_id,
    &a.Spots,
    &a.Activity_type,
    &a.Room,
    &a.Speaker,
    &a.Topic,
    &a.Description,
    &a.Time,
    &a.Day,
    )

  if err != nil {
    if err == sql.ErrNoRows {
      return a, fmt.Errorf("No activity found with id: %v\n", id)
    }
    return a, fmt.Errorf("could not retrieve activity: %v\n", err)
  }

  return a, nil
}

func CreateActivity(a Activity) (int, error) {
  query := `
  INSERT INTO activities
  (spots, activity_type, room, speaker, topic, description, time, day)
  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
  RETURNING id
  `
  var id int
  err := DB.QueryRow(query, a.Spots, a.Activity_type, a.Room, a.Speaker, a.Topic, a.Description, a.Time, a.Day).Scan(&id)
  if err != nil {
    return 0, fmt.Errorf("could not create activity: %v", err)
  }
  return id, nil
}

func (a Activity) String() string {
  return fmt.Sprintf("id: %v | spots: %v | day: %v | time: %v\nroom: %v | type: %v\nspeaker: %v | topic %v\ndescription: %v",
    a.Activity_id,
    a.Spots,
    a.Day,
    a.Time,
    a.Room,
    a.Activity_type,
    a.Speaker,
    a.Topic,
    a.Description,
    )
}

func SignupUserForActivity(userUUID string, activityID int) (bool, error) {
  tx, err := DB.Begin()
  if err != nil {
    return false, err
  }
  defer tx.Rollback()

  query := `
  SELECT COUNT(*)
  FROM registrations
  WHERE user_id = $1 AND activity_id = $2
  `

  var existingRegistration int
  err = tx.QueryRow(query, userUUID, activityID).Scan(&existingRegistration)
  if err != nil {
    return false, err
  }
  if existingRegistration > 0 {
    return false, errors.New("user is already signed up for this activity")
  }

  var activityDay, activitySpots int
  err = tx.QueryRow("SELECT day, spots FROM activities WHERE id = $1", activityID).Scan(&activityDay, &activitySpots)
  if err != nil {
    return false, err
  }

  query = `
  SELECT COUNT(*) FROM registrations r
  JOIN activities a ON r.activity_id = a.id
  WHERE r.user_id = $1 AND a.day = $2
  `

  var conflictingActivities int
  err = tx.QueryRow(query, userUUID, activityDay).Scan(&conflictingActivities)
  if err != nil {
    return false, err
  }
  if conflictingActivities > 0 {
    return false, errors.New("user already has an activity on this day")
  }

  if activitySpots <= 0 {
    return false, errors.New("no spots available")
  }

  query = `
  INSERT INTO registrations (user_id, activity_id) 
  VALUES ($1, $2)
  `
  _, err = tx.Exec(query, userUUID, activityID)
  if err != nil {
    return false, err
  }

  query = `
  UPDATE activities
  SET spots = spots - 1
  WHERE id = $1
  `
  _, err = tx.Exec(query, activityID)
  if err != nil {
    return false, err
  }

  err = tx.Commit()
  if err != nil {
    return false, err
  }

  return true, nil
}

func UnregisterUserFromActivity(userUUID string, activityID int) error {
  tx, err := DB.Begin()
  if err != nil {
    return err
  }
  defer tx.Rollback()

  query := `
  SELECT EXISTS(
  SELECT 1 FROM registrations
  WHERE user_id = $1 AND activity_id = $2
  )
  `

  var registrationExists bool
  err = tx.QueryRow(query, userUUID, activityID).Scan(&registrationExists)
  if err != nil {
    return err
  }
  if !registrationExists {
    return errors.New("user is not registered for this activity")
  }

  query = `
  DELETE FROM registrations
  WHERE user_id = $1 AND activity_id = $2
  `

  _, err = tx.Exec(query, userUUID, activityID)
  if err != nil {
    return err
  }

  query = `
  UPDATE activities 
  SET spots = spots + 1 
  WHERE id = $1
  `

  _, err = tx.Exec(query, activityID)
  if err != nil {
    return err
  }

  err = tx.Commit()
  if err != nil {
    return err
  }

  return nil
}

func GetUserActivities(userUUID string) ([]Activity, error) {
  query := `
  SELECT a.id, a.activity_type, a.room, a.speaker, a.topic, a.description, a.time, a.day, a.spots
  FROM activities a
  JOIN registrations r ON a.id = r.activity_id
  WHERE r.user_id = $1
  ORDER BY a.day, a.time
  `

  rows, err := DB.Query(query, userUUID)
  if err != nil {
    return nil, err
  }
  defer rows.Close()

  var activities []Activity
  for rows.Next() {
    var a Activity
    err := rows.Scan(&a.Activity_id, &a.Activity_type, &a.Room, &a.Speaker, &a.Topic, &a.Description, &a.Time, &a.Day, &a.Spots)
    if err != nil {
      return nil, err
    }
    activities = append(activities, a)
  }

  if err = rows.Err(); err != nil {
    return nil, err
  }

  return activities, nil
}
