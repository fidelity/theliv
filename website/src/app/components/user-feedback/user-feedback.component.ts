import { HttpErrorResponse } from '@angular/common/http';
import { Component, OnInit, Inject } from '@angular/core';
import { MatDialogRef, MAT_DIALOG_DATA} from '@angular/material/dialog';
import { KubernetesService } from 'src/app/services/kubernetes.service';

export interface FeedbackData {
  feedback: string;
}

@Component({
  selector: 'app-feedback',
  templateUrl: './user-feedback.component.html',
  styleUrls: ['./user-feedback.component.scss']
})
export class UserFeedbackComponent implements OnInit {

  complete = false;
  completeMsg = '';
  completeError = false;

  constructor(
    public dialogRef: MatDialogRef<UserFeedbackComponent>,
    @Inject(MAT_DIALOG_DATA) public data: FeedbackData,
    private kubeService: KubernetesService,
  ) {}

  ngOnInit(): void {}

  onNoClick(): void {
    this.dialogRef.close();
  }

  onSubmit(): void {
    this.complete = true;
    console.log(this.data.feedback);
    this.kubeService.postUserFeedback(this.data.feedback).subscribe(
      (res: any) => {
        if (res) {
          console.log(res);
        }
        this.completeError = false;
        this.completeMsg = 'Your feedback has been received. Thank you for helping us improve!';
      },
      (err: any) => {
        console.log('Post Feedback Error: ', err);
        this.completeError = true;
        this.completeMsg = 'Error occurred when posting feedback: ' + err.status + ' ' + err.statusText;
      }
    );
  }
}
