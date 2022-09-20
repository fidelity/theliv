/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
import { Component, OnInit } from '@angular/core';
import { IconProp } from '@fortawesome/fontawesome-svg-core';
import { faEnvelope, faTimes } from '@fortawesome/free-solid-svg-icons';

@Component({
  selector: 'app-feedback',
  templateUrl: './feedback.component.html',
  styleUrls: ['./feedback.component.scss']
})
export class FeedbackComponent implements OnInit {
  faEnvelope = faEnvelope as IconProp;
  faClose = faTimes as IconProp;
  isShowFeedback = false;

  constructor() { }

  ngOnInit(): void {
  }

  showFeedback() :void {
    this.isShowFeedback = !this.isShowFeedback;
  }
}
