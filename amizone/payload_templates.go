package amizone

const (
	facultyFeedbackTpl = `__RequestVerificationToken={{.VerificationToken}}&CourseType={{.CourseType}}&clsCourseFaculty.iDetId={{.DepartmentId}}&clsCourseFaculty.iFacultyStaffId={{.FacultyId}}&clsCourseFaculty.iSRNO={{.SerialNumber}}&FeedbackRating%5B0%5D.iAspectId=1&FeedbackRating%5B0%5D.Rating={{.Set__Rating}}&FeedbackRating%5B1%5D.iAspectId=2&FeedbackRating%5B1%5D.Rating={{.Set__Rating}}&FeedbackRating%5B2%5D.iAspectId=3&FeedbackRating%5B2%5D.Rating={{.Set__Rating}}&FeedbackRating%5B3%5D.iAspectId=4&FeedbackRating%5B3%5D.Rating={{.Set__Rating}}&FeedbackRating%5B4%5D.iAspectId=5&FeedbackRating%5B4%5D.Rating={{.Set__Rating}}&FeedbackRating%5B5%5D.iAspectId=6&FeedbackRating%5B5%5D.Rating={{.Set__Rating}}&FeedbackRating%5B6%5D.iAspectId=7&FeedbackRating%5B6%5D.Rating={{.Set__Rating}}&FeedbackRating%5B7%5D.iAspectId=8&FeedbackRating%5B7%5D.Rating={{.Set__Rating}}&FeedbackRating%5B8%5D.iAspectId=9&FeedbackRating%5B8%5D.Rating={{.Set__Rating}}&FeedbackRating%5B9%5D.iAspectId=10&FeedbackRating%5B9%5D.Rating={{.Set__Rating}}&FeedbackRating%5B10%5D.iAspectId=11&FeedbackRating%5B10%5D.Rating={{.Set__Rating}}&FeedbackRating%5B11%5D.iAspectId=12&FeedbackRating%5B11%5D.Rating={{.Set__Rating}}&FeedbackRating%5B12%5D.iAspectId=13&FeedbackRating%5B12%5D.Rating={{.Set__Rating}}&FeedbackRating%5B13%5D.iAspectId=14&FeedbackRating%5B13%5D.Rating={{.Set__Rating}}&FeedbackRating%5B14%5D.iAspectId=15&FeedbackRating%5B14%5D.Rating={{.Set__Rating}}&FeedbackRating%5B15%5D.iAspectId=16&FeedbackRating%5B15%5D.Rating={{.Set__Rating}}&FeedbackRating%5B16%5D.iAspectId=17&FeedbackRating%5B16%5D.Rating={{.Set__Rating}}&FeedbackRating%5B17%5D.iAspectId=18&FeedbackRating%5B17%5D.Rating={{.Set__Rating}}&FeedbackRating%5B18%5D.iAspectId=19&FeedbackRating%5B18%5D.Rating={{.Set__Rating}}&FeedbackRating%5B19%5D.iAspectId=20&FeedbackRating%5B19%5D.Rating={{.Set__Rating}}&FeedbackRating%5B20%5D.iAspectId=21&FeedbackRating%5B20%5D.Rating={{.Set__Rating}}&FeedbackRating%5B21%5D.iAspectId=22&FeedbackRating%5B21%5D.Rating={{.Set__Rating}}&FeedbackRating%5B22%5D.iAspectId=23&FeedbackRating%5B22%5D.Rating={{.Set__Rating}}&FeedbackRating%5B23%5D.iAspectId=24&FeedbackRating%5B23%5D.Rating={{.Set__Rating}}&FeedbackRating%5B24%5D.iAspectId=25&FeedbackRating%5B24%5D.Rating={{.Set__Rating}}&FeedbackRating_Q1Rating={{.Set__QRating}}&FeedbackRating_Q2Rating={{.Set__QRating}}&FeedbackRating_Q3Rating={{.Set__QRating}}&FeedbackRating_Comments={{.Set__Comment}}&X-Requested-With=XMLHttpRequest`
)