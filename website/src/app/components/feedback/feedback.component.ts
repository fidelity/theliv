/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
import { Component, Input, OnInit } from '@angular/core';
import { IconProp } from '@fortawesome/fontawesome-svg-core';
import { faEnvelope, faTimes, faCodeBranch, faCommentAlt } from '@fortawesome/free-solid-svg-icons';
import { KubernetesService } from 'src/app/services/kubernetes.service';

@Component({
  selector: 'app-feedback',
  templateUrl: './feedback.component.html',
  styleUrls: ['./feedback.component.scss']
})
export class FeedbackComponent implements OnInit {
  faEnvelope = faEnvelope as IconProp;
  faClose = faTimes as IconProp;
  faCodeBranch = faCodeBranch as IconProp;
  faCommentAlt = faCommentAlt as IconProp;
  isShowFeedback = false;
  @Input() configInfo: any;

  constructor(private kubeService: KubernetesService) { }

  ngOnInit(): void {
  }

  showFeedback() :void {
    this.isShowFeedback = !this.isShowFeedback;
  }
}
